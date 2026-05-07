require 'net/http'
require 'json'
require 'uri'

class WebhookSender
  Result = Struct.new(:success, :status, :body, :error, keyword_init: true)

  def send_webhook(backend_url:, endpoint:, shop_domain:, payload:, topic:)
    uri = URI("#{backend_url}#{endpoint}")
    http = Net::HTTP.new(uri.host, uri.port)
    http.use_ssl = uri.scheme == 'https'
    http.open_timeout = 5
    http.read_timeout = 10

    req = Net::HTTP::Post.new(uri.path)
    req['Content-Type'] = 'application/json'
    req['X-Shopify-Shop-Domain'] = shop_domain
    req['X-Shopify-Topic'] = topic
    req.body = payload.to_json

    response = http.request(req)
    Result.new(
      success: response.code.to_i < 400,
      status: response.code.to_i,
      body: response.body
    )
  rescue => e
    Result.new(success: false, status: 0, error: e.message)
  end

  # Build payload for each webhook type
  def self.build_payload(type, shop:, subscription_gid: nil, plan_name: nil, status: nil)
    case type
    when 'installed'
      {
        id: extract_shop_id(shop['gid']),
        myshopify_domain: shop['domain']
      }
    when 'uninstalled'
      {
        id: extract_shop_id(shop['gid']),
        name: shop['name'],
        myshopify_domain: shop['domain']
      }
    when 'subscriptions'
      {
        admin_graphql_api_id: subscription_gid || "gid://shopify/AppSubscription/#{rand(100000..999999)}",
        name: plan_name || 'Pro',
        status: status || 'ACTIVE'
      }
    when 'billing-failure'
      {
        admin_graphql_api_id: subscription_gid || "gid://shopify/AppSubscription/#{rand(100000..999999)}",
        status: 'FROZEN'
      }
    end
  end

  def self.extract_shop_id(gid)
    gid.to_s.split('/').last.to_i
  end

  def self.endpoint_for(type)
    "/webhooks/shopify/#{type}"
  end

  def self.topic_for(type)
    case type
    when 'installed' then 'app/installed'
    when 'uninstalled' then 'app/uninstalled'
    when 'subscriptions' then 'app_subscriptions/update'
    when 'billing-failure' then 'subscription_billing_attempts/failure'
    end
  end
end
