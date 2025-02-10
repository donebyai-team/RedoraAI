UPDATE conversations
SET summary          = :summary,
    recording_url    = :recording_url,
    call_duration    = :call_duration,
    end_of_call_reason = :end_of_call_reason,
    call_status        = :call_status,
    next_scheduled_at = :next_scheduled_at,
    call_messages = :call_messages,
    ai_decision = :ai_decision,
    external_id    = :external_id
WHERE id = :id;