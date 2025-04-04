import { FC, ReactNode } from 'react'

type ImageLoaderProps = {
  src: string
  width: number
  quality?: number
}

type ImageLoader = (p: ImageLoaderProps) => string

type PlaceholderValue = 'blur' | 'empty' | `data:image/${string}`
type OnLoadingComplete = (img: HTMLImageElement) => void

interface StaticImageData {
  src: string
  height: number
  width: number
  blurDataURL?: string
  blurWidth?: number
  blurHeight?: number
}
interface StaticRequire {
  default: StaticImageData
}

type StaticImport = StaticRequire | StaticImageData

export type ImageProps = Omit<
  React.DetailedHTMLProps<React.ImgHTMLAttributes<HTMLImageElement>, HTMLImageElement>,
  'height' | 'width' | 'loading' | 'ref' | 'alt' | 'src' | 'srcSet'
> & {
  src: string | StaticImport
  alt: string
  width?: number | `${number}` | undefined
  height?: number | `${number}` | undefined
  fill?: boolean | undefined
  loader?: ImageLoader | undefined
  quality?: number | `${number}` | undefined
  priority?: boolean | undefined
  loading?: 'eager' | 'lazy' | undefined
  placeholder?: PlaceholderValue | undefined
  blurDataURL?: string | undefined
  unoptimized?: boolean | undefined
  overrideSrc?: string | undefined
  onLoadingComplete?: OnLoadingComplete | undefined
  layout?: string | undefined
  objectFit?: string | undefined
  objectPosition?: string | undefined
  lazyBoundary?: string | undefined
  lazyRoot?: string | undefined
} & React.RefAttributes<HTMLImageElement | null>

/// The `ImageComponent` defines the TypeScript typings for the shared `Image` component that is used
/// as an abstraction between the Browser and the extensions.
//
// The typings are defined such that `next/image` can be used in the Browser and standard
// `img` tags can be used in the extensions.
export type ImageComponent = React.ForwardRefExoticComponent<ImageProps>

export const BrowserImage: FC<{ children: ReactNode; alt: string | undefined } & Record<string, unknown>> = ({
  children,
  alt,
  ...rest
}) => (
  <img alt={alt} {...filterNonBrowserProps(rest)}>
    {children}
  </img>
)

/// The `filterNonBrowserProps` function filters out the props that are not supported by the Browser
// but still pass in the "common" component and which might have an impact of Next rendering.
//
// The list is non-exhaustive and should be updated as needed.
const filterNonBrowserProps = (props: Record<string, unknown>): Record<string, unknown> =>
  Object.fromEntries(Object.entries(props).filter(([key]) => !['priority', 'loader'].includes(key)))
