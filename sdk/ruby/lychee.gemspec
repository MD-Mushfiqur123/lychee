Gem::Specification.new do |s|
  s.name        = "lychee"
  s.version     = "0.2.0"
  s.summary     = "Ruby client for Lychee — universal LLM runtime"
  s.description = "Official Ruby SDK for Lychee, the universal local LLM runtime and orchestration layer. Chat, generate, pull models, get embeddings, and compose multi-model pipelines."
  s.authors     = ["Lychee Tech"]
  s.email       = "hello@lychee.dev"
  s.homepage    = "https://github.com/MD-Mushfiqur123/lychee"
  s.license     = "MIT"
  s.files       = Dir.chdir(File.expand_path(__dir__)) do
    Dir["lib/**/*.rb"] + ["README.md", "license"]
  end
  s.require_paths = ["lib"]
  s.required_ruby_version = ">= 2.7"
  s.metadata = {
    "source_code_uri" => "https://github.com/MD-Mushfiqur123/lychee/tree/main/sdk/ruby",
    "bug_tracker_uri" => "https://github.com/MD-Mushfiqur123/lychee/issues",
    "documentation_uri" => "https://md-mushfiqur123.github.io/lychee-docs/"
  }
end
