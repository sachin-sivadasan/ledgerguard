require 'time'
require 'securerandom'

class TransactionResolver
  REV_SHARE = 0.20 # Shopify takes 20%

  def initialize(data_store)
    @store = data_store
    @cache = {} # org_id → { transactions: [...], generated_at: Time }
  end

  # Invalidate cached transactions for a persona (call after data changes)
  def invalidate_cache(org_id)
    @cache.delete(org_id.to_s)
  end

  # Resolve transactions query with pagination and date filters
  # Returns Shopify Partner API compatible response
  def resolve(org_id, first: 100, after: nil, created_at_min: nil, created_at_max: nil, app_id: nil)
    transactions = generate_transactions(org_id, created_at_min, created_at_max)

    # Filter by app if specified (real Shopify API returns only the requested app's data)
    if app_id
      transactions = transactions.select { |t| t[:node]['app'] && t[:node]['app']['id'] == app_id }
      # Re-assign cursors after filtering
      transactions.each_with_index { |t, i| t[:cursor] = "cursor_#{i}" }
    end

    # Handle cursor-based pagination
    start_index = 0
    if after
      start_index = transactions.index { |t| t[:cursor] == after }
      start_index = start_index ? start_index + 1 : transactions.length
    end

    page = transactions[start_index, first] || []
    has_next = (start_index + first) < transactions.length

    {
      'data' => {
        'transactions' => {
          'edges' => page.map { |t|
            {
              'cursor' => t[:cursor],
              'node' => t[:node]
            }
          },
          'pageInfo' => {
            'hasNextPage' => has_next
          }
        }
      }
    }
  end

  # Generate transactions for app discovery (no date filter, returns unique apps)
  def resolve_discovery(org_id, first: 100)
    resolve(org_id, first: first)
  end

  private

  def generate_transactions(org_id, created_at_min, created_at_max)
    cache_key = org_id.to_s
    no_date_filter = created_at_min.nil? && created_at_max.nil?

    # Return cached full-range results if available
    if no_date_filter && @cache[cache_key]
      return @cache[cache_key][:transactions]
    end

    transactions = []
    now = Time.now

    min_time = created_at_min ? Time.parse(created_at_min) : (now - 365 * 86400)
    max_time = created_at_max ? Time.parse(created_at_max) : now

    apps = @store.apps(org_id)
    shops = @store.shops(org_id)
    subs = @store.subscriptions(org_id)
    usage = @store.usage_charges(org_id)

    # Use seeded RNG for deterministic output per org
    rng = Random.new(org_id.to_i * 7919)

    # Generate subscription transactions (monthly charges)
    subs.each_with_index do |sub, idx|
      app = apps[sub['app_index']] rescue nil
      shop = shops[sub['shop_index']] rescue nil
      next unless app && shop

      price = sub['price'].to_f
      charge_id = "gid://partners/AppSubscriptionCharge/#{900000 + idx}"

      # Deterministic "last charge" offset per subscription
      last_charge_offset = 1 + (rng.rand(25))

      # Generate monthly charges going back from now
      if sub['status'] == 'active'
        # Active: charges every 30 days up to now
        charge_date = now - (last_charge_offset * 86400)
        12.times do |month|
          t = charge_date - (month * 30 * 86400)
          break if t < min_time
          next if t > max_time

          transactions << build_subscription_sale(
            id: idx * 100 + month,
            app: app,
            shop: shop,
            amount: price,
            charge_id: charge_id,
            created_at: t,
            billing_interval: sub['interval']
          )
        end
      elsif sub['status'] == 'frozen'
        # Frozen: charges stopped N days ago
        days_ago = sub['last_charge_days_ago'] || 45
        last_charge = now - (days_ago * 86400)
        8.times do |month|
          t = last_charge - (month * 30 * 86400)
          break if t < min_time
          next if t > max_time

          transactions << build_subscription_sale(
            id: idx * 100 + month,
            app: app,
            shop: shop,
            amount: price,
            charge_id: charge_id,
            created_at: t,
            billing_interval: sub['interval']
          )
        end
      elsif sub['status'] == 'cancelled'
        # Cancelled: a few old charges
        days_ago = sub['last_charge_days_ago'] || 100
        last_charge = now - (days_ago * 86400)
        4.times do |month|
          t = last_charge - (month * 30 * 86400)
          break if t < min_time
          next if t > max_time

          transactions << build_subscription_sale(
            id: idx * 100 + month,
            app: app,
            shop: shop,
            amount: price,
            charge_id: charge_id,
            created_at: t,
            billing_interval: sub['interval']
          )
        end
      end
    end

    # Generate usage charges (monthly, aligned to subscription billing cycle)
    usage.each_with_index do |uc, idx|
      app = apps[uc['app_index']] rescue nil
      shop = shops[uc['shop_index']] rescue nil
      next unless app && shop

      min_amt, max_amt = uc['amount_range']

      # Usage charges are billed monthly. started_months_ago controls how far back.
      started_months_ago = uc['started_months_ago'] || 12
      earliest_usage = now - (started_months_ago * 30 * 86400)
      effective_min = [min_time, earliest_usage].max

      charge_offset = 1 + rng.rand(25)
      charge_date = now - (charge_offset * 86400)
      started_months_ago.times do |month|
        t = charge_date - (month * 30 * 86400)
        break if t < effective_min
        next if t > max_time

        amount = (min_amt + rng.rand * (max_amt - min_amt)).round(2)
        transactions << build_usage_sale(
          id: 500000 + idx * 100 + month,
          app: app,
          shop: shop,
          amount: amount,
          created_at: t
        )
      end
    end

    # Sort by date descending
    transactions.sort_by! { |t| t[:node]['createdAt'] }.reverse!

    # Assign cursors
    transactions.each_with_index do |t, i|
      t[:cursor] = "cursor_#{i}"
    end

    # Cache full-range results
    if no_date_filter
      @cache[cache_key] = { transactions: transactions, generated_at: Time.now }
    end

    transactions
  end

  def build_subscription_sale(id:, app:, shop:, amount:, charge_id:, created_at:, billing_interval:)
    gross = amount
    net = (gross * (1 - REV_SHARE)).round(2)
    {
      cursor: nil,
      node: {
        'id' => "gid://partners/AppSubscriptionSale/#{id}",
        '__typename' => 'AppSubscriptionSale',
        'createdAt' => created_at.utc.iso8601,
        'netAmount' => { 'amount' => net.to_s, 'currencyCode' => 'USD' },
        'grossAmount' => { 'amount' => gross.to_s, 'currencyCode' => 'USD' },
        'shopifyFee' => { 'amount' => (gross - net).round(2).to_s, 'currencyCode' => 'USD' },
        'chargeId' => charge_id,
        'billingInterval' => billing_interval,
        'app' => { 'id' => app['id'], 'name' => app['name'] },
        'shop' => { 'myshopifyDomain' => shop['domain'], 'name' => shop['name'], 'id' => shop['gid'] }
      }
    }
  end

  def build_usage_sale(id:, app:, shop:, amount:, created_at:)
    gross = amount
    net = (gross * (1 - REV_SHARE)).round(2)
    {
      cursor: nil,
      node: {
        'id' => "gid://partners/AppUsageSale/#{id}",
        '__typename' => 'AppUsageSale',
        'createdAt' => created_at.utc.iso8601,
        'netAmount' => { 'amount' => net.to_s, 'currencyCode' => 'USD' },
        'grossAmount' => { 'amount' => gross.to_s, 'currencyCode' => 'USD' },
        'shopifyFee' => { 'amount' => (gross - net).round(2).to_s, 'currencyCode' => 'USD' },
        'app' => { 'id' => app['id'], 'name' => app['name'] },
        'shop' => { 'myshopifyDomain' => shop['domain'], 'name' => shop['name'], 'id' => shop['gid'] }
      }
    }
  end
end
