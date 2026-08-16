import React, { useState, useMemo } from 'react'
import { useUploadJobs } from '@/hooks/useDashboard'
import Card from '@/components/ui/card'
import Button from '@/components/ui/button'
import { DataTable, createColumnHelper } from '@/components/ui/DataTable'
import type { ColumnDef } from '@tanstack/react-table'
import type { UploadJob } from '@/types'
import { Loader2, AlertCircle, UploadCloud, CheckCircle2, Clock } from 'lucide-react'

const statusBadge = (status: string) => {
  const icon = status === 'COMPLETED' ? <CheckCircle2 className="h-3 w-3 mr-1" /> :
    status === 'PENDING' ? <Clock className="h-3 w-3 mr-1" /> :
    status === 'FAILED' ? <AlertCircle className="h-3 w-3 mr-1" /> : null
  const cls =
    status === 'COMPLETED' ? 'bg-green-100 text-green-800' :
    status === 'FAILED' ? 'bg-red-100 text-red-800' :
    status === 'PROCESSING' ? 'bg-blue-100 text-blue-800' :
    'bg-yellow-100 text-yellow-800'
  return (
    <span className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${cls}`}>
      {icon}
      {status}
    </span>
  )
}

const UploadJobs: React.FC = () => {
  const [status, setStatus] = useState('')
  const { data: jobs = [], isLoading, isError, refetch } = useUploadJobs(status || undefined)

  const columns: ColumnDef<UploadJob, any>[] = useMemo(() => {
    const helper = createColumnHelper<UploadJob>()
    return [
      helper.accessor('status', {
        header: 'Status',
        cell: (info) => statusBadge(info.getValue()),
      }),
      helper.accessor('object_key', {
        header: 'Object Key',
        cell: (info) => (
          <span className="font-mono text-gray-600 truncate max-w-xs block">{info.getValue()}</span>
        ),
      }),
      helper.accessor('attempt_count', {
        header: 'Attempts',
      }),
      helper.accessor('next_attempt_at', {
        header: 'Next Attempt',
        cell: (info) => info.getValue() ? new Date(info.getValue()).toLocaleString() : '-',
      }),
      helper.accessor('last_error', {
        header: 'Last Error',
        cell: (info) => (
          <span className="text-red-600 truncate max-w-xs block" title={info.getValue() || ''}>
            {info.getValue() || '-'}
          </span>
        ),
      }),
    ]
  }, [])

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-semibold text-gray-900">Upload Jobs</h1>
          <p className="text-sm text-gray-600">Monitor outgoing media uploads retries and failures</p>
        </div>
        <Button variant="secondary" size="sm" onClick={() => refetch()}>
          Refresh
        </Button>
      </div>

      <Card className="p-4 mb-6">
        <select
          className="form-control max-w-[200px]"
          value={status}
          onChange={(e) => setStatus(e.target.value)}
        >
          <option value="">All statuses</option>
          <option value="PENDING">Pending</option>
          <option value="PROCESSING">Processing</option>
          <option value="COMPLETED">Completed</option>
          <option value="FAILED">Failed</option>
        </select>
      </Card>

      {isLoading ? (
        <div className="flex justify-center py-12">
          <Loader2 className="h-8 w-8 animate-spin text-primary-500" />
        </div>
      ) : isError ? (
        <div className="text-center py-12 text-red-500">
          <AlertCircle className="h-12 w-12 mx-auto mb-3" />
          <p>Failed to load upload jobs.</p>
          <Button variant="primary" size="sm" className="mt-4" onClick={() => refetch()}>
            Retry
          </Button>
        </div>
      ) : jobs.length === 0 ? (
        <Card className="p-12 text-center">
          <UploadCloud className="h-12 w-12 text-gray-300 mx-auto mb-3" />
          <p className="text-gray-600">No upload jobs found.</p>
        </Card>
      ) : (
        <DataTable data={jobs} columns={columns} emptyMessage="No upload jobs found" />
      )}
    </div>
  )
}

export default UploadJobs
