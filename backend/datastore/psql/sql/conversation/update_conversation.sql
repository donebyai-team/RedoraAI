UPDATE conversations
SET summary          = :summary,
    recording_url    = :recording_url,
    call_duration    = :call_duration,
    end_of_call_reason = :end_of_call_reason,
    call_status        = :call_status,
    next_scheduled_at = :next_scheduled_at,
    external_id    = :external_id
WHERE id = :id;