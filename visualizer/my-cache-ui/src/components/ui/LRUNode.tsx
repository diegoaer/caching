import React, { memo } from 'react';
import { Handle, Position } from '@xyflow/react';

export default memo(({ data }: {
    data: {
        key: string;
        value: any;
        ttl?: number;
        isFirst?: boolean;
        isLast?: boolean;
        prev?: string;
        next?: string;
    }
}) => {
    return (
        <div className="rounded-md border border-gray-300 bg-white p-3 shadow-md text-sm text-black">
            <Handle type="target" position={Position.Left} isConnectable={data.isFirst} />
            <div className="text-sm text-gray-700">
                Key: {data.key}
                <br />
                Value: {String(data.value)}
            </div>
            {data.ttl != null && <div className="text-xs text-red-400">TTL: {data.ttl}s</div>}
            {!data.isLast && <Handle type="source" position={Position.Right} isConnectable={false} className="bg-white border border-gray-400" />}
        </div>
    );
});
