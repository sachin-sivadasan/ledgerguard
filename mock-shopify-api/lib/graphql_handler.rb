require_relative 'transaction_resolver'
require_relative 'app_events_resolver'

class GraphQLHandler
  def initialize(data_store)
    @transaction_resolver = TransactionResolver.new(data_store)
    @events_resolver = AppEventsResolver.new(data_store)
    @store = data_store
  end

  # Invalidate transaction cache for a persona (call after data changes)
  def invalidate_transaction_cache(org_id)
    @transaction_resolver.invalidate_cache(org_id)
  end

  def handle(org_id, query, variables = {})
    query = query.to_s.strip

    # Route based on query pattern
    if query.include?('transactions')
      handle_transactions(org_id, query, variables)
    elsif query.include?('events')
      handle_events(org_id, query, variables)
    else
      error_response("Unknown query pattern")
    end
  end

  private

  def handle_transactions(org_id, query, variables)
    # Extract parameters from query or variables
    first = extract_int(query, 'first') || variables['first'] || 100
    after = extract_string(query, 'after') || variables['after']
    created_at_min = extract_string(query, 'createdAtMin') || variables['createdAtMin']
    created_at_max = extract_string(query, 'createdAtMax') || variables['createdAtMax']
    app_id = extract_gid(query, 'appId') || variables['appId']

    # Check if this is a discovery query (no date filters)
    if created_at_min.nil? && created_at_max.nil? && !query.include?('createdAtMin')
      @transaction_resolver.resolve_discovery(org_id, first: first)
    else
      @transaction_resolver.resolve(
        org_id,
        first: first,
        after: after,
        created_at_min: created_at_min,
        created_at_max: created_at_max,
        app_id: app_id
      )
    end
  end

  def handle_events(org_id, query, variables)
    app_id = extract_gid(query, 'id') || variables['appId']
    shop_id = extract_string(query, 'shopId') || variables['shopId']
    types = extract_types(query) || variables['types']
    first = extract_int(query, 'first') || variables['first'] || 100
    after = extract_string(query, 'after') || variables['after']

    # If no app_id found, use the first app for this persona
    if app_id.nil?
      apps = @store.apps(org_id)
      app_id = apps.first['id'] if apps.any?
    end

    return error_response("No app found") unless app_id

    @events_resolver.resolve(
      org_id,
      app_id: app_id,
      shop_id: shop_id,
      types: types,
      first: first,
      after: after
    )
  end

  # Extract integer parameter from GraphQL query string
  def extract_int(query, param)
    match = query.match(/#{param}\s*:\s*(\d+)/i)
    match ? match[1].to_i : nil
  end

  # Extract string parameter (quoted) from GraphQL query string
  def extract_string(query, param)
    match = query.match(/#{param}\s*:\s*"([^"]+)"/i)
    match ? match[1] : nil
  end

  # Extract GID from query (e.g., app(id: "gid://partners/App/123"))
  def extract_gid(query, param)
    match = query.match(/#{param}\s*:\s*"(gid:\/\/[^"]+)"/i)
    match ? match[1] : nil
  end

  # Extract types array from query (e.g., types: [RELATIONSHIP_INSTALLED, RELATIONSHIP_UNINSTALLED])
  def extract_types(query)
    match = query.match(/types\s*:\s*\[([^\]]+)\]/i)
    return nil unless match
    match[1].split(',').map(&:strip)
  end

  def error_response(message)
    { 'errors' => [{ 'message' => message }] }
  end
end
