import { FC, ReactNode, createContext, useRef } from 'react'

export type DomContextType = {
  registerDomRef: (path: string, field: HTMLElement) => void
  getDomRef: (path: string) => HTMLElement | undefined
}

const defaultProvider: DomContextType = {
  registerDomRef: () => {},
  getDomRef: () => undefined
}

export const DOMContext = createContext<DomContextType>(defaultProvider)

type Props = {
  children: ReactNode
}

export const DOMProvider: FC<Props> = ({ children }) => {
  const attributeRefs = useRef<{ [key: string]: HTMLElement }>({})

  const registerDomRef = (path: string, field: HTMLElement) => {
    attributeRefs.current[path] = field
  }

  const getDomRef = (path: string) => {
    return attributeRefs.current[path]
  }

  const values = {
    registerDomRef,
    getDomRef
  }

  return <DOMContext.Provider value={values}>{children}</DOMContext.Provider>
}
