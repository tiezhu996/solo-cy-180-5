// 通用数据表格组件，跨页面复用。
import EmptyState from './EmptyState'
interface Column<T> {
  key: string
  title: string
  render: (row: T) => React.ReactNode
}

interface DataTableProps<T> {
  columns: Column<T>[]
  rows: T[]
  rowKey: (row: T) => number | string
  loading?: boolean
  emptyText?: string
}

export default function DataTable<T>({ columns, rows, rowKey, loading, emptyText = '暂无数据' }: DataTableProps<T>) {
  if (loading) {
    return <div className="table-loading">加载中…</div>
  }
  if (rows.length === 0) {
    return <EmptyState title={emptyText} description="" />
  }
  return (
    <div className="table-wrapper">
      <table className="data-table">
        <thead>
          <tr>
            {columns.map((col) => (
              <th key={col.key}>{col.title}</th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.map((row) => (
            <tr key={rowKey(row)}>
              {columns.map((col) => (
                <td key={col.key}>{col.render(row)}</td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
