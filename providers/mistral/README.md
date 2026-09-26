# Mistral Provider

This directory contains the Mistral provider implementation for the Fantasy library.

## Overview

The Mistral provider enables interaction with Mistral AI's API, supporting:
- Chat completions (non-streaming and streaming)
- Tool/function calling
- Structured outputs (if supported)
- Error handling

## API Documentation

See the [Mistral API Documentation](https://docs.mistral.ai/) for details on:
- Authentication
- Endpoints
- Request/response formats
- Streaming
- Tool/function calling
- Error handling

## Implementation Status

- [ ] Provider registration
- [ ] Language model implementation
- [ ] Authentication
- [ ] Error handling
- [ ] Streaming support
- [ ] Tool/function calling
- [ ] Structured outputs
- [ ] Tests
- [ ] Documentation

## Files

- `mistral.go`: Main provider implementation
- `language_model.go`: Language model implementation
- `error.go`: Error handling
- `provider_options.go`: Provider-specific options
- `call_useragent.go`: User-agent handling
