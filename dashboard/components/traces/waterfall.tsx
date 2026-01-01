'use client';

import { cn } from '@/lib/utils';

interface Span {
  id: string;
  name: string;
  startTime: number;
  duration: number;
  status: string;
}

interface TraceWaterfallProps {
  spans: Span[];
  totalDuration: number;
}

const colors = [
  'bg-indigo-500',
  'bg-blue-500',
  'bg-green-500',
  'bg-yellow-500',
  'bg-purple-500',
  'bg-pink-500',
  'bg-cyan-500',
];

export function TraceWaterfall({ spans, totalDuration }: TraceWaterfallProps) {
  return (
    <div className="space-y-2">
      {/* Timeline header */}
      <div className="flex items-center text-xs text-gray-500 mb-4">
        <div className="w-48 flex-shrink-0" />
        <div className="flex-1 flex justify-between">
          <span>0ms</span>
          <span>{Math.round(totalDuration / 4)}ms</span>
          <span>{Math.round(totalDuration / 2)}ms</span>
          <span>{Math.round((totalDuration * 3) / 4)}ms</span>
          <span>{totalDuration}ms</span>
        </div>
      </div>

      {/* Spans */}
      {spans.map((span, index) => {
        const left = (span.startTime / totalDuration) * 100;
        const width = (span.duration / totalDuration) * 100;
        const color = colors[index % colors.length];

        return (
          <div key={span.id} className="flex items-center gap-4 group">
            <div className="w-48 flex-shrink-0 text-right pr-4">
              <span className="text-sm font-medium text-gray-700 truncate block">
                {span.name}
              </span>
            </div>
            <div className="flex-1 relative h-8 bg-gray-100 rounded">
              <div
                className={cn(
                  'absolute h-full rounded transition-all',
                  color,
                  span.status === 'error' ? 'bg-red-500' : ''
                )}
                style={{
                  left: `${left}%`,
                  width: `${Math.max(width, 1)}%`,
                }}
              />
              <div
                className="absolute top-0 bottom-0 flex items-center opacity-0 group-hover:opacity-100 transition-opacity"
                style={{ left: `${left + width + 1}%` }}
              >
                <span className="text-xs text-gray-600 bg-white px-1 rounded shadow whitespace-nowrap">
                  {span.duration}ms
                </span>
              </div>
            </div>
          </div>
        );
      })}

      {/* Legend */}
      <div className="flex items-center gap-6 pt-4 mt-4 border-t border-gray-200">
        <div className="flex items-center gap-2 text-xs text-gray-500">
          <div className="w-3 h-3 rounded bg-indigo-500" />
          <span>Gateway</span>
        </div>
        <div className="flex items-center gap-2 text-xs text-gray-500">
          <div className="w-3 h-3 rounded bg-blue-500" />
          <span>Auth</span>
        </div>
        <div className="flex items-center gap-2 text-xs text-gray-500">
          <div className="w-3 h-3 rounded bg-green-500" />
          <span>RBAC</span>
        </div>
        <div className="flex items-center gap-2 text-xs text-gray-500">
          <div className="w-3 h-3 rounded bg-yellow-500" />
          <span>Safety</span>
        </div>
        <div className="flex items-center gap-2 text-xs text-gray-500">
          <div className="w-3 h-3 rounded bg-purple-500" />
          <span>MCP</span>
        </div>
      </div>
    </div>
  );
}
