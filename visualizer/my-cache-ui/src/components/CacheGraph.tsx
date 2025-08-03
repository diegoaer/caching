import { useEffect, useState } from 'react';
import { ReactFlow, Position, useNodesState, useEdgesState, ReactFlowProvider } from '@xyflow/react';
import type { Node, Edge } from "@xyflow/react";

import LRUNode from './ui/LRUNode';
import { fetchCacheState, addToCache } from '@/lib/api';

const nodeTypes = {
    lruNode: LRUNode,
};

interface CacheEntry {
    key: string;
    value: any;
    prev?: string;
    next?: string;
    expiresAt?: string;
}

export default function CacheGraph() {
    const [addingNode, setAddingNode] = useState(false);
    const [connectingFrom, setConnectingFrom] = useState<string | null>(null);
    const [nodes, setNodes, onNodesChange] = useNodesState<Node>([]);
    const [edges, setEdges, onEdgesChange] = useEdgesState<Edge>([]);

    const updateGraph = async () => {
        fetchCacheState()
            .then(({ capacity, items }: { capacity: number, items: CacheEntry[] }) => {

                const newEdges: Edge[] = items
                    .filter((entry) => entry.next)
                    .map((entry) => ({
                        id: `${entry.key}->${entry.next}`,
                        source: entry.key,
                        target: entry.next!,
                        animated: true,
                        style: { stroke: '#888' }
                    }));
                setEdges(newEdges);

                setNodes((prevNodes: Node[]) => {
                    return items.map((entry, index) => {
                        const existingNode = prevNodes.find(node => node.id === entry.key);
                        return {
                            id: entry.key,
                            type: 'lruNode',
                            position: existingNode?.position || { x: index * 180, y: 100 },
                            data: {
                                key: entry.key,
                                value: entry.value,
                                ttl: entry.expiresAt ? Math.floor((new Date(entry.expiresAt).getTime() - Date.now()) / 1000) : undefined,
                                isFirst: !entry.prev,
                                isLast: !entry.next,
                                prev: entry.prev,
                                next: entry.next,
                            },
                        };
                    });
                });
            })
    }

    useEffect(() => {
        if (!addingNode) {
            updateGraph();
        }
        const interval = setInterval(updateGraph, 5000); // Refresh every 5 seconds
        return () => clearInterval(interval);
    }, []);


    const onConnectStart = (_: unknown, params: { nodeId: string | null }) => {
        if (params.nodeId) {
            setConnectingFrom(params.nodeId);
        }
    };

    const onConnectEnd = async (pointerEvent: MouseEvent | TouchEvent) => {
        // If the connection ends on empty space
        const target = pointerEvent.target as HTMLElement;
        const droppedOnCanvas = target?.classList.contains("react-flow__pane");

        if (connectingFrom && droppedOnCanvas) {
            // Show prompt or modal to add a new key
            setAddingNode(true);
            const newKey = prompt("New key?");
            const newValue = prompt("New value?");
            if (newKey && newValue) {
                addToCache(newKey, newValue).then(updateGraph);
            }
            setConnectingFrom(null);
            setAddingNode(false);
        }

        await updateGraph(); // Refresh graph
    };

    return (
        <div style={{ width: '100%', height: '500px' }}>
            <ReactFlowProvider>
                <ReactFlow
                    nodes={nodes}
                    edges={edges}
                    onNodesChange={onNodesChange}
                    onEdgesChange={onEdgesChange}
                    nodeTypes={nodeTypes}
                    onConnectStart={onConnectStart}
                    onConnectEnd={onConnectEnd}
                    fitView
                />
            </ReactFlowProvider>
        </div>
    );
}
