import { FC, useContext } from 'react'
import { useElementRegistryContext } from '../context/ElementRegistryContext'
import { ImageProps } from '../bridge/Image'
import { ExecutionRuntimeContext, ExecutionRuntimeStore } from '../context/ExecutionRuntimeContext'

type Props = Omit<ImageProps, 'src'> & {
  imageKey: keyof ExecutionRuntimeStore['imageSourceRegistry']
}

// This component must be used within `ui-core` when rendering images as the actual image
// source differs based on the runtime environment.
//
// The `imageKey` prop is used to select the image source from the `imageSourceRegistry`, refer
// to @loadlogc/ui-core/context/ExecutionRuntimeContext for more information.
export const Image: FC<Props> = props => {
  const { Image: BaseImage } = useElementRegistryContext()
  const imageSourceStore = useContext(ExecutionRuntimeContext)

  const src = imageSourceStore.imageSourceRegistry[props.imageKey]
  const updatedProps = { ...props, src }

  // @ts-ignore We are removing the `imageKey` prop from the props passed to the `BaseImage` component because it's not required by ImageProps
  delete updatedProps['imageKey']

  return <BaseImage {...updatedProps} />
}
