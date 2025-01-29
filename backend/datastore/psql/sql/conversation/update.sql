UPDATE conversations
SET summary          = :summary,
    recording_url    = :recording_url,
    call_duration    = :call_duration,
    status           = :status,
    "external_id"    = :external_id
WHERE id = :id;