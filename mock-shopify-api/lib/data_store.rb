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
      cancelled: subs.count { |s| s['status'] == 'cancelled' }
    }
  end
end
