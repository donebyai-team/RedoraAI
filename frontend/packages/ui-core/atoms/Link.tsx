import { LinkProps } from '../bridge/Link'
import { FC } from 'react'
import { useElementRegistryContext } from '../context/ElementRegistryContext'

export const Link: FC<LinkProps> = props => {
  const { Link: BaseLink } = useElementRegistryContext()

  return <BaseLink {...props} />
}
