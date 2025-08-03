import React, { useCallback, useEffect, useState, useRef } from 'react';
import {
    ReactFlow,
    useNodesState,
    useEdgesState,
    useReactFlow,
    type Node,
    type Edge,
    type FinalConnectionState,
} from '@xyflow/react';

import LRUNode from '@/components/ui/LRUNode';
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
    const reactFlowWrapper = useRef(null);
    const [addingNode, setAddingNode] = useState(false);
    const [nodes, setNodes, onNodesChange] = useNodesState<Node[]>([]);
    const [edges, setEdges, onEdgesChange] = useEdgesState<Edge[]>([]);
    const { screenToFlowPosition } = useReactFlow();

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
            });
    }

    // 👇 Run once on mount
    useEffect(() => {
        updateGraph();
    }, []);

    useEffect(() => {
        if (addingNode) return;

        const interval = setInterval(updateGraph, 5000); // Refresh every 5 seconds

        return () => clearInterval(interval);
    }, [addingNode]);


    const onConnectStart = (_: unknown, params: { nodeId: string | null }) => {
        if (params.nodeId) {
            setAddingNode(true);
        }
    };

    const onConnectEnd = useCallback(
        (event: React.MouseEvent | React.TouchEvent, connectionState: FinalConnectionState) => {
            // when a connection is dropped on the pane it's not valid
            if (!connectionState.isValid) {
                // we need to remove the wrapper bounds, in order to get the correct position
                const newKey = prompt("New key?");
                const newValue = prompt("New value?");

                if (newKey && newValue) {
                    setNodes((prevNodes: Node[]) => {
                        const { clientX, clientY } =
                            'changedTouches' in event ? event.changedTouches[0] : event;
                        const position = screenToFlowPosition({
                            x: clientX,
                            y: clientY,
                        });
                        // If the node already exists, we update it
                        const existingNode = prevNodes.find(node => node.id === newKey);
                        const newNode: Node = {
                            id: newKey,
                            type: 'lruNode',
                            position,
                            data: {
                                key: newKey,
                                value: newValue,
                                ttl: undefined,
                                isFirst: false,
                                isLast: false,
                                prev: null,
                                next: null,
                            },
                        };
                        if (existingNode) {
                            return prevNodes.map(
                                node => node.id === newKey ? newNode : node
                            );
                        } else {
                            return [...prevNodes, newNode];
                        }
                    });
                    setEdges((prevEdges: Edge[]) => {
                        // We need to create an edge where the node was previously connected
                        const replacedEdge = {
                            source: null,
                            target: null,
                        }
                        const filteredEdges = prevEdges.map(edge => {
                            // If the edge is connected to the new node, we remove it
                            if (edge.source === newKey || edge.target === newKey) {
                                if (edge.source === newKey) {
                                    replacedEdge.target = edge.target;
                                } else {
                                    replacedEdge.source = edge.source;
                                }
                                return null;  // If the node already exists, we remove the previous edges
                            }
                            return edge;
                        });
                        return [
                            ...filteredEdges,
                            {
                                id: `${newKey}->${connectionState.fromNode.id}`,
                                source: newKey,
                                target: connectionState.fromNode.id,
                                animated: true,
                                style: { stroke: '#888' }
                            },
                            replacedEdge.source && replacedEdge.target ? {
                                id: `${replacedEdge.source}->${replacedEdge.target}`,
                                source: replacedEdge.source,
                                target: replacedEdge.target,
                                animated: true,
                                style: { stroke: '#888' }
                            } : null
                        ].filter(edge => edge !== null); // Remove any null edges
                    });
                    setAddingNode(false);
                    addToCache(newKey, newValue);
                } else {
                    setAddingNode(false);
                }
            }
        },
        [screenToFlowPosition],
    );
    return (
        <div
            style={{ width: '100%', height: '500px' }}
            className="wrapper"
            ref={reactFlowWrapper}
        >
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
        </div>
    );
}
