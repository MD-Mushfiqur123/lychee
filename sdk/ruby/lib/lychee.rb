require 'net/http'
require 'json'
require 'uri'

# LycheeClient — Official Ruby SDK for Lychee
# Universal local LLM runtime and orchestration layer.
#
# @example Quick start
#   client = LycheeClient.new
#   resp = client.chat("gemma3", [{role: "user", content: "Hello!"}])
#   puts resp["message"]["content"]
#
class LycheeClient
  VERSION = "0.2.0"

  attr_reader :host, :timeout

  # @param host [String] Lychee server URL (default: http://localhost:11434)
  # @param timeout [Integer] HTTP request timeout in seconds (default: 120)
  def initialize(host = 'http://localhost:11434', timeout: 120)
    @host = host.chomp('/')
    @timeout = timeout
  end

  # ── Chat ──────────────────────────────────────────────────────────

  # Chat with a model using a message list (conversation format).
  #
  # @param model [String] model name
  # @param messages [Array<Hash>] array of {role:, content:} messages
  # @param stream [Boolean] enable streaming response
  # @param options [Hash] additional model parameters
  # @return [Hash, Enumerator] response hash or streaming enumerator
  def chat(model, messages, stream: false, **options)
    payload = { model: model, messages: messages, stream: stream }.merge(options)
    uri = URI("#{@host}/api/chat")
    if stream
      stream_post(uri, payload) { |chunk| yield chunk if block_given? }
    else
      post_json(uri, payload)
    end
  end

  # ── Generate ──────────────────────────────────────────────────────

  # Generate text from a model (single-turn prompt).
  #
  # @param model [String] model name
  # @param prompt [String] the prompt text
  # @param stream [Boolean] enable streaming
  # @param options [Hash] additional model parameters
  # @return [Hash, Enumerator] response hash or streaming enumerator
  def generate(model, prompt, stream: false, **options)
    payload = { model: model, prompt: prompt, stream: stream }.merge(options)
    uri = URI("#{@host}/api/generate")
    if stream
      stream_post(uri, payload) { |chunk| yield chunk if block_given? }
    else
      post_json(uri, payload)
    end
  end

  # ── Models ────────────────────────────────────────────────────────

  # List all locally installed models.
  #
  # @return [Hash] response with models array
  def list_models
    get_json(URI("#{@host}/api/tags"))
  end

  # Show details for a specific model.
  #
  # @param model [String] model name
  # @return [Hash] model details
  def show_model(model)
    payload = { name: model }
    post_json(URI("#{@host}/api/show"), payload)
  end

  # Delete a locally installed model.
  #
  # @param model [String] model name to delete
  # @return [Hash] response
  def delete_model(model)
    uri = URI("#{@host}/api/delete")
    req = Net::HTTP::Delete.new(uri)
    req.body = { name: model }.to_json
    req.content_type = 'application/json'
    do_request(uri, req)
  end

  # List currently running models.
  #
  # @return [Hash] response with running models
  def list_running
    get_json(URI("#{@host}/api/ps"))
  end

  # ── Pull ──────────────────────────────────────────────────────────

  # Pull a model from HuggingFace.
  # Yields progress hashes when a block is given; returns the final
  # response otherwise.
  #
  # @param name [String] model name (e.g. "bartowski/Meta-Llama-3.1-8B-Instruct-GGUF")
  # @param insecure [Boolean] allow insecure connections
  # @return [Hash, Enumerator] pull response or progress enumerator
  def pull_model(name, insecure: false)
    payload = { name: name, stream: true, insecure: insecure }
    uri = URI("#{@host}/api/pull")
    if block_given?
      stream_post(uri, payload) { |chunk| yield chunk }
    else
      stream_post(uri, payload)
    end
  end

  # ── Embeddings ────────────────────────────────────────────────────

  # Get an embedding vector for a single input string.
  #
  # @param model [String] embedding model name
  # @param input [String] text to embed
  # @return [Array<Float>] embedding vector
  def embeddings(model, input)
    payload = { model: model, input: input }
    resp = post_json(URI("#{@host}/api/embed"), payload)
    resp["embeddings"]&.first || []
  end

  # Get embedding vectors for multiple inputs.
  #
  # @param model [String] embedding model name
  # @param inputs [Array<String>] texts to embed
  # @return [Array<Array<Float>>] array of embedding vectors
  def embeddings_batch(model, inputs)
    inputs.map { |input| embeddings(model, input) }
  end

  # ── Compose (Multi-Model DAG Pipelines) ───────────────────────────

  # Execute a multi-model DAG pipeline.
  #
  # @param input [String] initial input
  # @param steps [Array<Hash>] pipeline steps
  # @param stream [Boolean] enable streaming
  # @return [Hash] compose result
  def compose(input, steps, stream: false)
    payload = { input: input, steps: steps, stream: stream }
    post_json(URI("#{@host}/api/compose"), payload)
  end

  # ── Structured Output ─────────────────────────────────────────────

  # Generate JSON conforming to a schema, with automatic retries.
  #
  # @param model [String] model name
  # @param prompt [String] input prompt
  # @param schema [Hash] JSON schema to validate against
  # @param max_retries [Integer] maximum retry attempts (default: 3)
  # @param options [Hash] additional model parameters
  # @return [Hash, Array] parsed structured output
  def structured(model, prompt, schema, max_retries: 3, **options)
    payload = {
      model: model,
      prompt: prompt,
      schema: schema,
      max_retries: max_retries
    }.merge(options)
    post_json(URI("#{@host}/api/structured"), payload)
  end

  # ── Server Info ───────────────────────────────────────────────────

  # Get server version string.
  #
  # @return [String] version
  def version
    resp = get_json(URI("#{@host}/api/version"))
    resp["version"] || "unknown"
  rescue
    "unknown"
  end

  # Check if the server is reachable.
  #
  # @return [Boolean] true if server responds
  def health
    get_json(URI("#{@host}/api/version"))
    true
  rescue
    false
  end

  # ── Internal HTTP Helpers ─────────────────────────────────────────

  private

  def get_json(uri)
    req = Net::HTTP::Get.new(uri)
    do_request(uri, req)
  end

  def post_json(uri, body_hash)
    req = Net::HTTP::Post.new(uri)
    req.body = body_hash.to_json
    req.content_type = 'application/json'
    do_request(uri, req)
  end

  def do_request(uri, req)
    http = Net::HTTP.new(uri.host, uri.port)
    http.use_ssl = (uri.scheme == 'https')
    http.open_timeout = @timeout
    http.read_timeout = @timeout

    res = http.request(req)

    raise "HTTP #{res.code}: #{res.body}" unless res.code.to_i < 400

    return {} if res.body.nil? || res.body.empty?
    JSON.parse(res.body)
  rescue JSON::ParserError
    {}
  end

  # Stream a POST request. Yields each parsed JSON line when a block is
  # given; returns an Enumerator otherwise.
  def stream_post(uri, body_hash)
    req = Net::HTTP::Post.new(uri)
    req.body = body_hash.to_json
    req.content_type = 'application/json'

    http = Net::HTTP.new(uri.host, uri.port)
    http.use_ssl = (uri.scheme == 'https')
    http.open_timeout = @timeout
    http.read_timeout = @timeout

    enumerator = Enumerator.new do |yielder|
      http.request(req) do |response|
        response.read_body do |chunk|
          chunk.each_line do |line|
            line = line.strip
            next if line.empty?
            begin
              obj = JSON.parse(line)
              yielder << obj
              yield obj if block_given?
            rescue JSON::ParserError
              # skip malformed lines in stream
            end
          end
        end
      end
    end

    if block_given?
      enumerator.each {} # consume to trigger yields
      nil
    else
      enumerator
    end
  end
end
