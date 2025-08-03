import CacheGraph from "@/components/CacheGraph";
import { ReactFlowProvider } from '@xyflow/react';


function App() {
  return (
    <div className="p-4 bg-white shadow-lg rounded-xl max-w-5xl mx-auto mt-8">
      <h1 className="text-2xl font-bold mb-4">LRU Cache Visualizer</h1>
      <ReactFlowProvider>
        <CacheGraph />
      </ReactFlowProvider>
    </div>
  );
}

export default App;
