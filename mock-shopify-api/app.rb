require 'sinatra/base'
require 'json'
require_relative 'lib/data_store'
require_relative 'lib/graphql_handler'
require_relative 'lib/webhook_sender'

class MockShopifyAPI < Sinatra::Base
  set :port, 4000
  set :bind, '0.0.0.0'
  set :views, File.join(File.dirname(__FILE__), 'views')
  enable :sessions

  configure do
    set :data_store, DataStore.new
    set :graphql_handler, GraphQLHandler.new(settings.data_store)
    set :webhook_sender, WebhookSender.new
    set :webhook_history, []
  end

  # ─── GraphQL Endpoint ──────────────────────────────────────────────

  # POST /:org_id/api/2025-07/graphql.json
  post '/:org_id/api/2025-07/graphql.json' do
    content_type :json
    org_id = params[:org_id]

    unless settings.data_store.persona(org_id)
      status 404
      return { errors: [{ message: "Unknown org_id: #{org_id}" }] }.to_json
    end

    body = JSON.parse(request.body.read) rescue {}
    query = body['query'] || ''
    variables = body['variables'] || {}

    result = settings.graphql_handler.handle(org_id, query, variables)
    result.to_json
  end

  # ─── Admin UI ──────────────────────────────────────────────────────

  get '/admin' do
    @personas = settings.data_store.personas
    erb :admin
  end

  get '/admin/personas/:org_id' do
    org_id = params[:org_id]
    @persona = settings.data_store.persona(org_id)
    halt 404, "Persona not found" unless @persona

    @apps = settings.data_store.apps(org_id)
    @shops = settings.data_store.shops(org_id)
    @subscriptions = settings.data_store.subscriptions(org_id)
    @events = settings.data_store.events(org_id)
    @stats = settings.data_store.stats(org_id)
    @org_id = org_id
    erb :persona
  end

  # Add a shop
  post '/admin/personas/:org_id/shops' do
    org_id = params[:org_id]
    data = settings.data_store.persona_data(org_id)
    halt 404, "Persona not found" unless data

    # Add the shop
    shop_index = data['shops'].length
    data['shops'] << {
      'domain' => params[:domain],
      'name' => params[:name],
      'gid' => "gid://shopify/Shop/#{rand(900000..999999)}"
    }

    # Auto-create subscription linking shop to app
    data['subscriptions'] << {
      'app_index' => params[:app_index].to_i,
      'shop_index' => shop_index,
      'plan_name' => params[:plan_name] || 'Pro',
      'price' => params[:price] || '29.99',
      'interval' => 'EVERY_30_DAYS',
      'status' => 'active'
    }

    # Auto-create install event
    data['events'] << {
      'type' => 'RELATIONSHIP_INSTALLED',
      'app_index' => params[:app_index].to_i,
      'shop_index' => shop_index,
      'days_ago' => 0
    }

    settings.data_store.save_persona(org_id)
    redirect "/admin/personas/#{org_id}"
  end

  # Add a subscription
  post '/admin/personas/:org_id/subscriptions' do
    org_id = params[:org_id]
    data = settings.data_store.persona_data(org_id)
    halt 404, "Persona not found" unless data

    sub = {
      'app_index' => params[:app_index].to_i,
      'shop_index' => params[:shop_index].to_i,
      'plan_name' => params[:plan_name],
      'price' => params[:price],
      'interval' => params[:interval] || 'EVERY_30_DAYS',
      'status' => params[:status] || 'active'
    }
    sub['last_charge_days_ago'] = params[:last_charge_days_ago].to_i if params[:last_charge_days_ago].to_s != ''

    data['subscriptions'] << sub
    settings.data_store.save_persona(org_id)
    redirect "/admin/personas/#{org_id}"
  end

  # Update subscription status
  post '/admin/personas/:org_id/subscriptions/:idx/status' do
    org_id = params[:org_id]
    data = settings.data_store.persona_data(org_id)
    halt 404, "Persona not found" unless data

    idx = params[:idx].to_i
    sub = data['subscriptions'][idx]
    halt 404, "Subscription not found" unless sub

    sub['status'] = params[:status]
    sub['last_charge_days_ago'] = params[:last_charge_days_ago].to_i if params[:last_charge_days_ago].to_s != ''

    settings.data_store.save_persona(org_id)
    redirect "/admin/personas/#{org_id}"
  end

  # Add an event
  post '/admin/personas/:org_id/events' do
    org_id = params[:org_id]
    data = settings.data_store.persona_data(org_id)
    halt 404, "Persona not found" unless data

    data['events'] << {
      'type' => params[:type],
      'app_index' => params[:app_index].to_i,
      'shop_index' => params[:shop_index].to_i,
      'days_ago' => params[:days_ago].to_i
    }
    settings.data_store.save_persona(org_id)
    redirect "/admin/personas/#{org_id}"
  end

  # Reload data from YAML files
  post '/admin/reload' do
    settings.data_store.reload
    settings.set :graphql_handler, GraphQLHandler.new(settings.data_store)
    redirect '/admin'
  end

  # ─── Webhook Tester ────────────────────────────────────────────────

  get '/admin/webhooks' do
    @personas = settings.data_store.personas
    @history = settings.webhook_history
    erb :webhooks
  end

  # API to get shops for a persona (AJAX)
  get '/admin/api/shops/:org_id' do
    content_type :json
    shops = settings.data_store.shops(params[:org_id])
    shops.map.with_index { |s, i| { index: i, domain: s['domain'], name: s['name'], gid: s['gid'] } }.to_json
  end

  # Send webhook
  post '/admin/webhooks/send' do
    org_id = params[:persona_org_id]
    shop_index = params[:shop_index].to_i
    webhook_type = params[:webhook_type]
    backend_url = params[:backend_url] || 'http://localhost:8080'
    status_value = params[:status_value]

    shops = settings.data_store.shops(org_id)
    shop = shops[shop_index]
    halt 400, "Shop not found" unless shop

    payload = WebhookSender.build_payload(
      webhook_type,
      shop: shop,
      plan_name: params[:plan_name],
      status: status_value
    )
    endpoint = WebhookSender.endpoint_for(webhook_type)
    topic = WebhookSender.topic_for(webhook_type)

    result = settings.webhook_sender.send_webhook(
      backend_url: backend_url,
      endpoint: endpoint,
      shop_domain: shop['domain'],
      payload: payload,
      topic: topic
    )

    # Save to history
    entry = {
      time: Time.now.strftime('%H:%M:%S'),
      type: webhook_type,
      shop: shop['domain'],
      status: result.status,
      success: result.success,
      error: result.error,
      payload: payload
    }
    settings.webhook_history.unshift(entry)
    settings.webhook_history.pop if settings.webhook_history.length > 20

    if request.xhr?
      content_type :json
      entry.to_json
    else
      redirect '/admin/webhooks'
    end
  end

  # ─── Health check ──────────────────────────────────────────────────

  get '/health' do
    content_type :json
    { status: 'ok', personas: settings.data_store.personas.keys }.to_json
  end

  run! if app_file == $0
end
