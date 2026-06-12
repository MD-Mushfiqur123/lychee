import test from 'node:test';
import assert from 'node:assert';
import { Lychee, LycheeError } from '../src/index.js';

test('Lychee Client Initialization', (t) => {
  const client = new Lychee();
  assert.strictEqual(client.baseUrl, 'http://localhost:11434');

  const clientCustom = new Lychee('http://192.168.1.50:11434/');
  assert.strictEqual(clientCustom.baseUrl, 'http://192.168.1.50:11434');
});

test('Lychee Client methods success flows', async (t) => {
  const client = new Lychee();
  const originalFetch = globalThis.fetch;
  
  t.after(() => {
    globalThis.fetch = originalFetch;
  });

  // Test Chat
  globalThis.fetch = async (url, options) => {
    assert.strictEqual(url, 'http://localhost:11434/api/chat');
    assert.strictEqual(options.method, 'POST');
    const body = JSON.parse(options.body);
    assert.strictEqual(body.model, 'gemma3');
    assert.strictEqual(body.messages[0].content, 'Hello');
    assert.strictEqual(body.stream, false);

    return {
      ok: true,
      json: async () => ({ message: { role: 'assistant', content: 'Hi!' } })
    };
  };
  const chatRes = await client.chat('gemma3', 'Hello');
  assert.strictEqual(chatRes.message.content, 'Hi!');

  // Test Generate
  globalThis.fetch = async (url, options) => {
    assert.strictEqual(url, 'http://localhost:11434/api/generate');
    const body = JSON.parse(options.body);
    assert.strictEqual(body.prompt, 'Write a poem');
    return {
      ok: true,
      json: async () => ({ response: 'Once upon a time...' })
    };
  };
  const genRes = await client.generate('gemma3', 'Write a poem');
  assert.strictEqual(genRes.response, 'Once upon a time...');

  // Test Compose
  globalThis.fetch = async (url, options) => {
    assert.strictEqual(url, 'http://localhost:11434/api/compose');
    const body = JSON.parse(options.body);
    assert.strictEqual(body.input, 'hello');
    assert.strictEqual(body.steps[0].model, 'gemma3');
    return {
      ok: true,
      json: async () => ({ output: 'composed output', results: [] })
    };
  };
  const compRes = await client.compose({
    input: 'hello',
    steps: [{ model: 'gemma3', prompt: 'test' }]
  });
  assert.strictEqual(compRes.output, 'composed output');

  // Test Messages (Anthropic compatible)
  globalThis.fetch = async (url, options) => {
    assert.strictEqual(url, 'http://localhost:11434/v1/messages');
    const body = JSON.parse(options.body);
    assert.strictEqual(body.max_tokens, 500);
    return {
      ok: true,
      json: async () => ({ content: [{ type: 'text', text: 'Anthropic reply' }] })
    };
  };
  const msgRes = await client.messages('gemma3', [{ role: 'user', content: 'hi' }], { maxTokens: 500 });
  assert.strictEqual(msgRes.content[0].text, 'Anthropic reply');

  // Test Chat Completions (OpenAI compatible)
  globalThis.fetch = async (url, options) => {
    assert.strictEqual(url, 'http://localhost:11434/v1/chat/completions');
    const body = JSON.parse(options.body);
    assert.strictEqual(body.temperature, 0.7);
    return {
      ok: true,
      json: async () => ({ choices: [{ message: { content: 'OpenAI reply' } }] })
    };
  };
  const compOpenAI = await client.chatCompletions('gemma3', [{ role: 'user', content: 'hi' }], { temperature: 0.7 });
  assert.strictEqual(compOpenAI.choices[0].message.content, 'OpenAI reply');

  // Test List Models
  globalThis.fetch = async (url, options) => {
    assert.strictEqual(url, 'http://localhost:11434/api/tags');
    return {
      ok: true,
      json: async () => ({ models: [{ name: 'gemma3' }, { name: 'llama3' }] })
    };
  };
  const models = await client.listModels();
  assert.strictEqual(models.length, 2);
  assert.strictEqual(models[0].name, 'gemma3');

  // Test Pull
  globalThis.fetch = async (url, options) => {
    assert.strictEqual(url, 'http://localhost:11434/api/pull');
    const body = JSON.parse(options.body);
    assert.strictEqual(body.model, 'gemma3');
    return {
      ok: true,
      json: async () => ({ status: 'success' })
    };
  };
  const pullRes = await client.pull('gemma3', { stream: false });
  assert.strictEqual(pullRes.status, 'success');

  // Test Show
  globalThis.fetch = async (url, options) => {
    assert.strictEqual(url, 'http://localhost:11434/api/show');
    const body = JSON.parse(options.body);
    assert.strictEqual(body.model, 'gemma3');
    return {
      ok: true,
      json: async () => ({ details: { parameter_size: '9B' } })
    };
  };
  const showRes = await client.show('gemma3');
  assert.strictEqual(showRes.details.parameter_size, '9B');

  // Test Structured
  globalThis.fetch = async (url, options) => {
    assert.strictEqual(url, 'http://localhost:11434/api/structured');
    const body = JSON.parse(options.body);
    assert.strictEqual(body.model, 'gemma3');
    assert.strictEqual(body.max_retries, 3);
    return {
      ok: true,
      json: async () => ({ output: '{"name":"Alice"}', valid: true, attempts: 1 })
    };
  };
  const structRes = await client.structured('gemma3', 'extract name', { type: 'object' });
  assert.strictEqual(structRes.valid, true);
  assert.strictEqual(structRes.attempts, 1);

  // Test List Conversations
  globalThis.fetch = async (url, options = {}) => {
    assert.strictEqual(url, 'http://localhost:11434/api/conversations?limit=50&offset=0');
    assert.strictEqual(options.method ?? 'GET', 'GET');
    return {
      ok: true,
      json: async () => ([{ id: 'conv-1', title: 'Chat 1' }])
    };
  };
  const convList = await client.listConversations();
  assert.strictEqual(convList.length, 1);
  assert.strictEqual(convList[0].id, 'conv-1');

  // Test Get Conversation
  globalThis.fetch = async (url, options = {}) => {
    assert.strictEqual(url, 'http://localhost:11434/api/conversations/conv-1');
    assert.strictEqual(options.method ?? 'GET', 'GET');
    return {
      ok: true,
      json: async () => ({ id: 'conv-1', messages: [] })
    };
  };
  const convGet = await client.getConversation('conv-1');
  assert.strictEqual(convGet.id, 'conv-1');

  // Test Delete Conversation
  globalThis.fetch = async (url, options = {}) => {
    assert.strictEqual(url, 'http://localhost:11434/api/conversations/conv-1');
    assert.strictEqual(options.method, 'DELETE');
    return {
      ok: true,
      text: async () => JSON.stringify({ status: 'deleted' })
    };
  };
  const convDel = await client.deleteConversation('conv-1');
  assert.strictEqual(convDel.status, 'deleted');

  // Test Create Route
  globalThis.fetch = async (url, options = {}) => {
    assert.strictEqual(url, 'http://localhost:11434/api/routes');
    assert.strictEqual(options.method, 'POST');
    const body = JSON.parse(options.body);
    assert.strictEqual(body.name, 'fast');
    return {
      ok: true,
      json: async () => ({ name: 'fast', strategy: 'round_robin' })
    };
  };
  const routeCreate = await client.createRoute('fast', [{ host: 'http://localhost:11434' }]);
  assert.strictEqual(routeCreate.name, 'fast');

  // Test List Routes
  globalThis.fetch = async (url, options = {}) => {
    assert.strictEqual(url, 'http://localhost:11434/api/routes');
    assert.strictEqual(options.method ?? 'GET', 'GET');
    return {
      ok: true,
      json: async () => ([{ name: 'fast', strategy: 'round_robin' }])
    };
  };
  const routeList = await client.listRoutes();
  assert.strictEqual(routeList.length, 1);
  assert.strictEqual(routeList[0].name, 'fast');

  // Test Delete Route
  globalThis.fetch = async (url, options = {}) => {
    assert.strictEqual(url, 'http://localhost:11434/api/routes/fast');
    assert.strictEqual(options.method, 'DELETE');
    return {
      ok: true,
      text: async () => JSON.stringify({ status: 'deleted' })
    };
  };
  const routeDel = await client.deleteRoute('fast');
  assert.strictEqual(routeDel.status, 'deleted');

  // Test Is Running
  globalThis.fetch = async (url, options) => {
    assert.strictEqual(url, 'http://localhost:11434/');
    return {
      ok: true,
      json: async () => ({})
    };
  };
  const isRun = await client.isRunning();
  assert.strictEqual(isRun, true);
});

test('Lychee Client Error Handling', async (t) => {
  const client = new Lychee();
  const originalFetch = globalThis.fetch;
  
  t.after(() => {
    globalThis.fetch = originalFetch;
  });

  globalThis.fetch = async (url, options) => {
    return {
      ok: false,
      status: 404,
      text: async () => 'model not found'
    };
  };

  await assert.rejects(
    async () => {
      await client.chat('nonexistent', 'hi');
    },
    (err) => {
      assert.ok(err instanceof LycheeError);
      assert.strictEqual(err.statusCode, 404);
      assert.match(err.message, /HTTP 404: model not found/);
      return true;
    }
  );
});
