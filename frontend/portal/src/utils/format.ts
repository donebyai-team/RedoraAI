// utils/format.ts

import { SourceTyeps } from '@/store/Source/sourceSlice'
import { Timestamp } from '@bufbuild/protobuf/wkt'
import { formatDistanceToNow } from 'date-fns'

export function getFormattedDate(timestamp: Timestamp | undefined): string {
  if (!timestamp) return 'N/A'

  try {
    const millis = Number(timestamp.seconds) * 1000
    const date = new Date(millis)
    const timeAgo = formatDistanceToNow(date, { addSuffix: true })
    return timeAgo
  } catch (e) {
    console.error('Date formatting error:', e)
    return 'N/A'
  }
}

export const getSubredditName = (list: SourceTyeps[], id: string) => {
  const name = list?.find(reddit => reddit.id === id)?.name ?? "N/A";
  return name;
};
