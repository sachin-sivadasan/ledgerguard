require 'yaml'
require 'json'

class DataStore
  attr_reader :personas

  def initialize(data_dir = File.join(File.dirname(__FILE__), '..', 'data'))
    @data_dir = data_dir
    @personas = {}
    load_all
  end

  def load_all
    index = YAML.load_file(File.join(@data_dir, 'personas.yml'))
    index['personas'].each do |org_id, meta|
      file_path = File.join(@data_dir, meta['file'])
      data = YAML.load_file(file_path)
      @personas[org_id] = {
        'org_id' => org_id,
        'name' => meta['name'],
        'description' => meta['description'],
        'file' => meta['file'],
        'data' => data
      }
      expand_generated(org_id) if data['generated']
      expand_usage_charges(org_id)
      backfill_missing_events(org_id)
    end
  end

  def reload
    @personas = {}
    load_all
  end

  def persona(org_id)
    @personas[org_id.to_s]
  end

  def persona_data(org_id)
    p = persona(org_id)
    p ? p['data'] : nil
  end

  def apps(org_id)
    data = persona_data(org_id)
    data ? (data['apps'] || []) : []
  end

  def shops(org_id)
    data = persona_data(org_id)
    data ? (data['shops'] || []) : []
  end

  def subscriptions(org_id)
    data = persona_data(org_id)
    data ? (data['subscriptions'] || []) : []
  end

  def usage_charges(org_id)
    data = persona_data(org_id)
    data ? (data['usage_charges'] || []) : []
  end

  def events(org_id)
    data = persona_data(org_id)
    data ? (data['events'] || []) : []
  end

  def save_persona(org_id)
    p = persona(org_id)
    return unless p

    file_path = File.join(@data_dir, p['file'])
    File.write(file_path, YAML.dump(p['data']))
  end

  # Summary stats for a persona
  def stats(org_id)
    subs = subscriptions(org_id)
    {
      apps: apps(org_id).length,
      shops: shops(org_id).length,
      subscriptions: subs.length,
      active: subs.count { |s| s['status'] == 'active' },
      frozen: subs.count { |s| s['status'] == 'frozen' },
      cancelled: subs.count { |s| s['status'] == 'cancelled' },
      usage_charges: usage_charges(org_id).length
    }
  end

  private

  # Expand a persona's `generated` block into shops and subscriptions.
  # Usage charges are expanded separately by expand_usage_charges.
  # Deterministic: uses seeded RNG so paginated API calls return consistent data.
  def expand_generated(org_id)
    data = persona_data(org_id)
    return unless data && data['generated']

    gen = data['generated']
    rng = Random.new(org_id.to_i)

    num_shops = gen['shops'] || 0
    distribution = gen['distribution'] || []
    status_mix = gen['status_mix'] || { 'active' => 0.75, 'frozen' => 0.15, 'cancelled' => 0.10 }

    shops = []
    subscriptions = []

    # Generate shops and subscriptions
    shop_index = 0
    distribution.each do |dist|
      app_idx = dist['app_index']
      count = (num_shops * dist['ratio']).round
      plans = dist['plans'] || []

      count.times do
        break if shop_index >= num_shops

        shops << {
          'domain' => "shop-#{shop_index}.myshopify.com",
          'name' => "Shop #{shop_index}",
          'gid' => "gid://shopify/Shop/5#{shop_index.to_s.rjust(6, '0')}"
        }

        # Pick plan based on pct distribution
        roll = rng.rand
        cumulative = 0.0
        chosen_plan = plans.last || { 'name' => 'Basic', 'price' => 9.99 }
        plans.each do |plan|
          cumulative += plan['pct']
          if roll <= cumulative
            chosen_plan = plan
            break
          end
        end

        # Pick status based on status_mix
        status_roll = rng.rand
        status = if status_roll < status_mix['active']
                   'active'
                 elsif status_roll < status_mix['active'] + status_mix['frozen']
                   'frozen'
                 else
                   'cancelled'
                 end

        sub = {
          'app_index' => app_idx,
          'shop_index' => shop_index,
          'plan_name' => chosen_plan['name'],
          'price' => chosen_plan['price'].to_s,
          'interval' => 'EVERY_30_DAYS',
          'status' => status
        }

        if status == 'frozen'
          sub['last_charge_days_ago'] = 31 + rng.rand(30)
        elsif status == 'cancelled'
          sub['last_charge_days_ago'] = 60 + rng.rand(120)
        end

        subscriptions << sub
        shop_index += 1
      end
    end

    data['shops'] = shops
    data['subscriptions'] = subscriptions

    # Generate events from subscriptions
    events = []
    evt_rng = Random.new(org_id.to_i * 17)
    subscriptions.each do |sub|
      si = sub['shop_index']
      ai = sub['app_index']
      status = sub['status']

      # Every shop gets an install event; every 10th is recent (deterministic for dev limit)
      install_days = (si % 10 == 0) ? evt_rng.rand(14) : 90 + evt_rng.rand(270)
      events << { 'type' => 'RELATIONSHIP_INSTALLED', 'app_index' => ai, 'shop_index' => si, 'days_ago' => install_days }

      # Active shops get a recent subscription charge
      if status == 'active'
        events << { 'type' => 'SUBSCRIPTION_CHARGE_ACCEPTED', 'app_index' => ai, 'shop_index' => si, 'days_ago' => evt_rng.rand(30) }
      elsif status == 'frozen'
        charge_days = sub['last_charge_days_ago'] || 45
        events << { 'type' => 'SUBSCRIPTION_CHARGE_FROZEN', 'app_index' => ai, 'shop_index' => si, 'days_ago' => charge_days }
      elsif status == 'cancelled'
        cancel_days = sub['last_charge_days_ago'] || 90
        events << { 'type' => 'SUBSCRIPTION_CHARGE_CANCELED', 'app_index' => ai, 'shop_index' => si, 'days_ago' => cancel_days }
        # Some cancelled shops also uninstalled
        if evt_rng.rand < 0.6
          # ~10% of uninstalls are recent (visible in "This Week"/"This Month")
          uninstall_days = evt_rng.rand < 0.1 ? evt_rng.rand(7) : [cancel_days - 2, 1].max
          events << { 'type' => 'RELATIONSHIP_UNINSTALLED', 'app_index' => ai, 'shop_index' => si, 'days_ago' => uninstall_days }
        end
      end
    end
    data['events'] = events

    data.delete('generated')
  end

  # Expand usage charge templates with shop_ratio into per-shop entries.
  # Runs for ALL personas. Entries without shop_ratio are left as-is.
  # Each expanded entry gets a started_months_ago (1-10) so transactions
  # don't all start from the same date.
  def expand_usage_charges(org_id)
    data = persona_data(org_id)
    return unless data

    usage_charges = data['usage_charges'] || []
    return if usage_charges.empty?

    # Check if any templates need expanding
    has_templates = usage_charges.any? { |uc| uc['shop_ratio'] }
    return unless has_templates

    rng = Random.new(org_id.to_i * 31)

    # Build app → shop indices mapping from subscriptions
    subs = data['subscriptions'] || []
    app_shop_indices = Hash.new { |h, k| h[k] = [] }
    subs.each do |sub|
      app_shop_indices[sub['app_index']] << sub['shop_index']
    end
    # Deduplicate (a shop could have multiple subs to same app)
    app_shop_indices.each { |_k, v| v.uniq! }

    expanded = []
    usage_charges.each do |uc|
      shop_ratio = uc['shop_ratio']
      if shop_ratio && shop_ratio > 0
        app_idx = uc['app_index']
        candidates = app_shop_indices[app_idx] || []
        next if candidates.empty?

        num = (candidates.length * shop_ratio).round
        selected = candidates.sample(num, random: rng)
        selected.each do |si|
          expanded << {
            'app_index' => app_idx,
            'shop_index' => si,
            'amount_range' => uc['amount_range'],
            'started_months_ago' => 1 + rng.rand(10)
          }
        end
      else
        # Explicit entry — add started_months_ago if missing
        uc['started_months_ago'] ||= 1 + rng.rand(10)
        expanded << uc
      end
    end

    data['usage_charges'] = expanded
  end

  # Backfill events for subscriptions that have no install event.
  # Hand-crafted persona YAMLs only define events for a few representative shops;
  # this ensures every subscription has at least an install event plus a
  # status-appropriate lifecycle event.
  def backfill_missing_events(org_id)
    data = persona_data(org_id)
    return unless data

    subs = data['subscriptions'] || []
    events = data['events'] || []
    return if subs.empty?

    # Build set of (app_index, shop_index) that already have an install event
    covered = Set.new
    events.each do |ev|
      if ev['type'] == 'RELATIONSHIP_INSTALLED'
        covered.add([ev['app_index'], ev['shop_index']])
      end
    end

    rng = Random.new(org_id.to_i * 23)
    new_events = []

    subs.each do |sub|
      key = [sub['app_index'], sub['shop_index']]
      next if covered.include?(key)

      ai = sub['app_index']
      si = sub['shop_index']
      status = sub['status']

      # Install event; every 10th is recent (deterministic for dev limit)
      install_days = (si % 10 == 0) ? rng.rand(14) : 90 + rng.rand(270)
      new_events << { 'type' => 'RELATIONSHIP_INSTALLED', 'app_index' => ai, 'shop_index' => si, 'days_ago' => install_days }

      # Status-appropriate lifecycle event
      case status
      when 'active'
        new_events << { 'type' => 'SUBSCRIPTION_CHARGE_ACCEPTED', 'app_index' => ai, 'shop_index' => si, 'days_ago' => rng.rand(30) }
      when 'frozen'
        charge_days = sub['last_charge_days_ago'] || 45
        new_events << { 'type' => 'SUBSCRIPTION_CHARGE_FROZEN', 'app_index' => ai, 'shop_index' => si, 'days_ago' => charge_days }
      when 'cancelled'
        cancel_days = sub['last_charge_days_ago'] || 90
        new_events << { 'type' => 'SUBSCRIPTION_CHARGE_CANCELED', 'app_index' => ai, 'shop_index' => si, 'days_ago' => cancel_days }
        if rng.rand < 0.6
          uninstall_days = rng.rand < 0.1 ? rng.rand(7) : [cancel_days - 2, 1].max
          new_events << { 'type' => 'RELATIONSHIP_UNINSTALLED', 'app_index' => ai, 'shop_index' => si, 'days_ago' => uninstall_days }
        end
      end
    end

    data['events'] = events + new_events unless new_events.empty?
  end
end
