import { useCallback, useLayoutEffect, useRef } from 'react'

type TimerHandle = ReturnType<typeof setTimeout>

// eslint-disable-next-line @typescript-eslint/ban-types
export function useDebounce<T extends Function>(callback: T, delay: number): T {
  const callbackRef = useRef(callback)

  useLayoutEffect(() => {
    callbackRef.current = callback
  })

  let timer: TimerHandle

  // @ts-ignore Parameters<T> doesn't work because Function is not accepted as input for Parameters
  const naiveDebounce = (func: T, delayMs: number, ...args: Parameters<T>) => {
    clearTimeout(timer)
    timer = setTimeout(() => {
      func(...args)
    }, delayMs)
  }

  return useCallback<T>(
    // @ts-ignore Parameters<T> doesn't work because Function is not accepted as input for Parameters
    (...args: Parameters<T>) => naiveDebounce(callbackRef.current, delay, ...args),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [delay]
  )
}
