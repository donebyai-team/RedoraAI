import { useCallback, useContext } from 'react'
import { ExecutionRuntime, ExecutionRuntimeContext, ExecutionRuntimeStore } from '../context/ExecutionRuntimeContext'

export const useExecutionRuntime = (): ExecutionRuntimeStore['runtime'] => {
  const store = useContext(ExecutionRuntimeContext)

  return store.runtime
}


export const useExecutionActions = (): ExecutionRuntimeStore['action'] => {
  const store = useContext(ExecutionRuntimeContext)

  return store.action
}


export const useExecutionRuntimeIsOneOf = (...runtimes: ExecutionRuntime[]): boolean => {
  const store = useContext(ExecutionRuntimeContext)

  return runtimes.includes(store.runtime)
}

export const useIsExecutionRuntimeInPortal = (): boolean => {
  const store = useContext(ExecutionRuntimeContext)

  return store.runtime == 'portal'
}

export const useIsExecutionRuntimeInExtension = (): boolean => {
  const store = useContext(ExecutionRuntimeContext)

  return ['chrome_extension', 'outlook_addin'].includes(store.runtime)
}

type ClassMapFunction = (base: string, details: Record<string, string>) => string

/**
 * Makes it easy to conditionally add classes based on the execution runtime in one compact
 * format:
 *
 * ```tsx
 * const classMap = useExecutionRuntimeClassMap()
 *
 * <div
 *   className={classMap('flex flex-col gap-2', {
 *     portal: 'border-l'
 *   })}
 * >
 * ```
 * @returns A `classMap` that dynamically adds classes based on the execution runtime.
 */
export const useExecutionRuntimeClassMap = (): ClassMapFunction => {
  const runtime = useExecutionRuntime()
  const mapper = useCallback(
    (base: string, details: Record<string, string>) => {
      const extra = details[runtime]
      if (!extra) {
        return base
      }

      return `${base} ${extra}`
    },
    [runtime]
  )

  return mapper
}
