import { useRef, useCallback } from 'react'
import { toast } from '@/components/ui/use-toast'
import { Lead, LeadStatus } from '@doota/pb/doota/core/v1/core_pb'
import { useAppDispatch } from '@/store/hooks'
import { portalClient } from '@/services/grpc'
import { setError, setIsLoading, setLeadList, setLeadStatusFilter } from '@/store/Lead/leadSlice'
import { DEFAULT_DATA_LIMIT } from '@/utils/constants'
import { DateRangeFilter } from '@doota/pb/doota/portal/v1/portal_pb'

export interface LeadFetchParams {
  status: LeadStatus | null
  relevancyScore?: number
  subReddit?: string
  dateRange?: DateRangeFilter
  pageCount?: number
  pageNo?: number
}

export interface FetchOptions {
  pageNo?: number
  shouldFallbackToCompletedLeads?: boolean
  fetchType?: 'initial' | 'pagination'
}

export interface UseLeadFetcherProps {
  relevancyScore?: number
  subReddit?: string
  dateRange?: DateRangeFilter
  leadStatusFilter: LeadStatus | null
  leadList: Lead[]
  setPageNo: (pageNo: number) => void
  setCounts?: (data: any) => void
  setHasMore: (hasMore: boolean) => void
  setIsFetchingMore: (isLoading: boolean) => void
}

export const useLeadFetcher = ({
  relevancyScore,
  subReddit,
  dateRange,
  leadStatusFilter,
  leadList,
  setPageNo,
  setCounts,
  setHasMore,
  setIsFetchingMore
}: UseLeadFetcherProps) => {
  const dispatch = useAppDispatch()
  const hasRunPriorityLoad = useRef(false)

  // gRPC wrapper
  const fetchLeadsFromServer = useCallback(async (params: LeadFetchParams) => {
    return await portalClient.getRelevantLeads({
      ...(params.relevancyScore && { relevancyScore: params.relevancyScore }),
      ...(params.subReddit && { subReddit: params.subReddit }),
      ...(params.status && { status: params.status }),
      ...(params.dateRange && { dateRange: params.dateRange }),
      pageCount: params.pageCount ?? DEFAULT_DATA_LIMIT,
      pageNo: params.pageNo ?? 1
    })
  }, [])

  // Common handler: Loading State
  const setLoadingState = (type: 'initial' | 'pagination', isLoading: boolean) => {
    type === 'initial' ? dispatch(setIsLoading(isLoading)) : setIsFetchingMore(isLoading)
  }

  // Common handler: Error
  const handleError = (err: any) => {
    const message = err?.response?.data?.message || err.message || 'Something went wrong'
    toast({ title: 'Error', description: message })
    dispatch(setError(message))
  }

  // Common handler: Success
  const handleSuccess = (response: Awaited<ReturnType<typeof fetchLeadsFromServer>>, pageNo: number) => {
    const newLeads = response?.leads ?? []
    const hasMore = newLeads.length === DEFAULT_DATA_LIMIT

    dispatch(setLeadList([...leadList, ...newLeads]))
    setCounts?.(response.analysis)
    setHasMore(hasMore)
    setPageNo(pageNo)
  }

  // Main Fetch Logic
  const fetchLeads = useCallback(
    async ({ pageNo = 1, shouldFallbackToCompletedLeads = false, fetchType = 'initial' }: FetchOptions) => {
      setLoadingState(fetchType, true)

      try {
        if (shouldFallbackToCompletedLeads && !hasRunPriorityLoad.current) {
          const priorityStatuses: LeadStatus[] = [LeadStatus.NEW, LeadStatus.COMPLETED]

          for (const status of priorityStatuses) {
            try {
              const response = await fetchLeadsFromServer({
                status,
                relevancyScore,
                subReddit,
                dateRange,
                pageNo
              })

              if ((response.leads ?? []).length > 0) {
                handleSuccess(response, pageNo)
                dispatch(setLeadStatusFilter(status))
                break
              }
            } catch (err: any) {
              handleError(err)
            }
          }

          hasRunPriorityLoad.current = true
        } else {
          const response = await fetchLeadsFromServer({
            status: leadStatusFilter,
            relevancyScore,
            subReddit,
            dateRange,
            pageNo
          })

          handleSuccess(response, pageNo)
        }
      } catch (err: any) {
        handleError(err)
      } finally {
        setLoadingState(fetchType, false)
      }
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [setLoadingState, fetchLeadsFromServer, relevancyScore, subReddit, dateRange, dispatch, leadStatusFilter]
  )

  return { fetchLeads }
}
