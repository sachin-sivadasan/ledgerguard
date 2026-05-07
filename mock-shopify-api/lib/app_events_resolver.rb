require 'time'

class AppEventsResolver
  def initialize(data_store)
    @store = data_store
  end

  # Resolve events for a specific app, optionally filtered by shop and/or event types
  def resolve(org_id, app_id:, shop_id: nil, types: nil, first: 100, after: nil)
    events = generate_events(org_id, app_id, shop_id: shop_id, types: types)

    # Cursor pagination
    start_index = 0
    if after
      start_index = events.index { |e| e[:cursor] == after }
      start_index = start_index ? start_index + 1 : events.length
    end

    page = events[start_index, first] || []
    has_next = (start_index + first) < events.length

    {
      'data' => {
        'app' => {
          'id' => app_id,
          'events' => {
            'edges' => page.map { |e|
              {
                'cursor' => e[:cursor],
                'node' => e[:node]
              }
            },
            'pageInfo' => {
              'hasNextPage' => has_next
            }
          }
        }
      }
    }
  end

  private

  def generate_events(org_id, app_id, shop_id: nil, types: nil)
    now = Time.now
    apps = @store.apps(org_id)
    shops = @store.shops(org_id)
    raw_events = @store.events(org_id)

    # Find app index from app_id
    app_index = apps.index { |a| a['id'] == app_id }
    return [] unless app_index

    events = []
    raw_events.each_with_index do |evt, idx|
      next unless evt['app_index'] == app_index

      shop = shops[evt['shop_index']] rescue nil
      next unless shop

      # Filter by shop if specified
      if shop_id
        next unless shop['gid'] == shop_id || shop['domain'] == shop_id
      end

      # Filter by types if specified
      if types && !types.empty?
        next unless types.include?(evt['type'])
      end

      occurred_at = now - (evt['days_ago'] * 86400)

      events << {
        cursor: nil,
        node: {
          'type' => evt['type'],
          'occurredAt' => occurred_at.utc.iso8601,
          'shop' => {
            'myshopifyDomain' => shop['domain'],
            'name' => shop['name'],
            'id' => shop['gid']
          }
        }
      }
    end

    # Sort by date descending
    events.sort_by! { |e| e[:node]['occurredAt'] }.reverse!

    # Assign cursors
    events.each_with_index { |e, i| e[:cursor] = "evt_cursor_#{i}" }

    events
  end
end
