UPDATE conversations
SET summary          = :summary,
    recording_url    = :recording_url,
    call_duration    = :call_duration,
    end_of_call_reason = :end_of_call_reason,
    status           = :status,
    external_id    = :external_id
WHERE id = :id;