import React, { useState } from 'react'
import {
  useReactTable,
  getCoreRowModel,
  getSortedRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  flexRender,
  createColumnHelper,
  type ColumnDef,
  type SortingState,
  type ColumnFiltersState,
} from '@tanstack/react-table'
import { ChevronUp, ChevronDown, ChevronsUpDown } from 'lucide-react'
import Card from '@/components/ui/card'
import Button from '@/components/ui/button'

export interface DataTableProps<TData> {
  data: TData[]
  columns: ColumnDef<TData, any>[]
  initialState?: {
    sorting?: SortingState
    pagination?: { pageSize?: number }
    columnVisibility?: Record<string, boolean>
  }
  renderActions?: (row: TData) => React.ReactNode
  emptyMessage?: string
  enableGlobalFilter?: boolean
}

export function DataTable<TData extends { id: string }>({
  data,
  columns,
  initialState,
  renderActions,
  emptyMessage = 'No data',
  enableGlobalFilter = true,
}: DataTableProps<TData>) {
  const [sorting, setSorting] = useState<SortingState>(initialState?.sorting ?? [])
  const [globalFilter, setGlobalFilter] = useState('')
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([])

  const table = useReactTable<TData>({
    data,
    columns,
    state: {
      sorting,
      globalFilter,
      columnFilters,
    },
    onSortingChange: setSorting,
    onGlobalFilterChange: setGlobalFilter,
    onColumnFiltersChange: setColumnFilters,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    initialState: {
      pagination: { pageSize: 10 },
      ...initialState,
    },
  })

  return (
    <Card className="overflow-hidden">
      {enableGlobalFilter && (
        <div className="p-3 border-b border-gray-200">
          <input
            className="form-control max-w-xs"
            placeholder="Search..."
            value={globalFilter}
            onChange={(e) => setGlobalFilter(e.target.value)}
          />
        </div>
      )}

      <div className="overflow-x-auto">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            {table.getHeaderGroups().map((headerGroup) => (
              <tr key={headerGroup.id}>
                {headerGroup.headers.map((header) => (
                  <th
                    key={header.id}
                    className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                    colSpan={header.colSpan}
                  >
                    {header.isPlaceholder ? null : header.column.getCanSort() ? (
                      <button
                        className="inline-flex items-center gap-1 hover:text-gray-900"
                        onClick={header.column.getToggleSortingHandler()}
                      >
                        {flexRender(header.column.columnDef.header, header.getContext())}
                        {{
                          asc: <ChevronUp className="h-3 w-3" />,
                          desc: <ChevronDown className="h-3 w-3" />,
                        }[header.column.getIsSorted() as string] ?? (
                          <ChevronsUpDown className="h-3 w-3 text-gray-300" />
                        )}
                      </button>
                    ) : (
                      flexRender(header.column.columnDef.header, header.getContext())
                    )}
                  </th>
                ))}
              </tr>
            ))}
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {table.getRowModel().rows.length === 0 ? (
              <tr>
                <td colSpan={columns.length} className="px-6 py-8 text-center text-gray-500">
                  {emptyMessage}
                </td>
              </tr>
            ) : (
              table.getRowModel().rows.map((row) => (
                <tr key={row.id} className="hover:bg-gray-50">
                  {row.getVisibleCells().map((cell) => (
                    <td key={cell.id} className="px-6 py-4 whitespace-nowrap text-sm text-gray-600">
                      {flexRender(cell.column.columnDef.cell, cell.getContext())}
                    </td>
                  ))}
                  {renderActions && (
                    <td className="px-6 py-4 whitespace-nowrap text-sm" onClick={(e) => e.stopPropagation()}>
                      {renderActions(row.original)}
                    </td>
                  )}
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {table.getPageCount() > 1 && (
        <div className="p-3 border-t border-gray-200 flex items-center justify-between">
          <span className="text-sm text-gray-500">
            Page {table.getState().pagination.pageIndex + 1} of {table.getPageCount()}
          </span>
          <div className="flex gap-2">
            <Button
              variant="secondary"
              size="sm"
              onClick={() => table.previousPage()}
              disabled={!table.getCanPreviousPage()}
            >
              Previous
            </Button>
            <Button
              variant="secondary"
              size="sm"
              onClick={() => table.nextPage()}
              disabled={!table.getCanNextPage()}
            >
              Next
            </Button>
          </div>
        </div>
      )}
    </Card>
  )
}

export { createColumnHelper }
