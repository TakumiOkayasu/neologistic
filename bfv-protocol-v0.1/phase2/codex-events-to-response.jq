([.[] | select(.type == "item.completed" and .item.type == "agent_message")] | last) as $message
| ([.[] | select(.type == "turn.completed")] | last) as $turn
| if $message == null then error("no completed agent_message event")
  elif $turn == null then error("no turn.completed event")
  else {
    raw: $message.item.text,
    usage: {
      input_tokens: ($turn.usage.input_tokens // null),
      cached_input_tokens: ($turn.usage.cached_input_tokens // null),
      cache_write_input_tokens: ($turn.usage.cache_write_input_tokens // null),
      output_tokens: ($turn.usage.output_tokens // null),
      reasoning_output_tokens: ($turn.usage.reasoning_output_tokens // null)
    }
  }
  end
