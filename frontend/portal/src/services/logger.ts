import pino from 'pino'

const logger = pino({
  level: process.env.NEXT_PUBLIC_LOG_LEVEL || inferDefaultLogLevel()
})

export const log = logger

function inferDefaultLogLevel() {
  if (process.env.NODE_ENV === 'development' && process.env.DEBUG_TEST == null) {
    return 'info'
  }

  return 'silent'
}
