// 分页 hook。
import { useCallback, useState } from 'react'

export function usePagination(initialPage = 1, initialPageSize = 10) {
  const [page, setPage] = useState(initialPage)
  const [pageSize, setPageSize] = useState(initialPageSize)
  const [total, setTotal] = useState(0)

  const reset = useCallback(() => {
    setPage(1)
  }, [])

  return {
    page,
    pageSize,
    total,
    setPage,
    setPageSize,
    setTotal,
    reset,
    totalPages: Math.max(1, Math.ceil(total / pageSize)),
  }
}
