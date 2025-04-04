import { createContext, FC, ReactNode, useContext } from 'react'
import { BrowserLink, LinkComponent } from '../bridge/Link'
import { BrowserImage, ImageComponent } from '../bridge/Image'

export type ElementRegistry = {
  Link: LinkComponent
  Image: ImageComponent
}

export const ElementRegistryContext = createContext<ElementRegistry>({
  // @ts-ignore - Weird TypeScript dependency resolution leading to multiple version of some library leading to incompatibility, ignoring for now
  Link: BrowserLink,
  // @ts-ignore - Weird TypeScript dependency resolution leading to multiple version of some library leading to incompatibility, ignoring for now
  Image: BrowserImage
})

export const useElementRegistryContext = (): ElementRegistry => useContext<ElementRegistry>(ElementRegistryContext)

type Props = {
  Link: LinkComponent
  Image: ImageComponent
  children: ReactNode
}

export const BaseElementRegistryProvider: FC<Props> = ({ Link, Image, children }) => {
  return (
    <ElementRegistryContext.Provider
      value={{
        Link: Link,
        Image: Image
      }}
    >
      {children}
    </ElementRegistryContext.Provider>
  )
}
