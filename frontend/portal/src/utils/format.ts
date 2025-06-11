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

export function getFreePlanDateStatus(expiredAt?: Timestamp | undefined): string {
  if (!expiredAt) return "";

  const expiryDate = new Date(Number(expiredAt.seconds) * 1000);
  const today = new Date();

  expiryDate.setHours(0, 0, 0, 0);
  today.setHours(0, 0, 0, 0);

  const diffInMs = expiryDate.getTime() - today.getTime();
  const diffInDays = Math.ceil(diffInMs / (1000 * 60 * 60 * 24));

  if (diffInDays > 0) {
    return `(${diffInDays} days left)`;
  } else if (diffInDays === 0) {
    return `(expires today)`;
  } else {
    return "";
  }
}

export function formatTimestampToReadableDate(timestamp?: Timestamp): string {
  if (!timestamp || !timestamp.seconds) return "No expiry";

  const millis = Number(timestamp.seconds) * 1000 + Number(timestamp.nanos || 0) / 1_000_000;
  const date = new Date(millis);

  return date.toLocaleDateString("en-US", {
    year: "numeric",
    month: "long",
    day: "numeric",
  });
};

export const getSubredditName = (list: SourceTyeps[], id: string) => {
  const name = list?.find(reddit => reddit.id === id)?.name ?? "N/A";
  return name;
};
