([.[] | select(.type == "item.completed" and .item.type == "agent_message")] | last) as $message
| ([.[] | select(.type == "turn.completed")] | last) as $turn
| if $message == null then error("no completed agent_message event")
  elif $turn == null then error("no turn.completed event")
  else {
    case_id: $case_id,
    candidate: "pipe-v1",
    trial: $trial,
    raw: $message.item.text,
    usage: {
      input_tokens: ($turn.usage.input_tokens // 0),
      cached_tokens: ($turn.usage.cached_input_tokens // 0),
      output_tokens: ($turn.usage.output_tokens // 0),
      reasoning_tokens: ($turn.usage.reasoning_output_tokens // 0)
    }
  }
  end
