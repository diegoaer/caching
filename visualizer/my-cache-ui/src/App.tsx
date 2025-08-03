import CacheGraph from "@/components/CacheGraph";


function App() {
  return (
    <div className="p-4 bg-white shadow-lg rounded-xl max-w-5xl mx-auto mt-8">
      <h1 className="text-2xl font-bold mb-4">LRU Cache Visualizer</h1>
      <CacheGraph />
    </div>
  );
}

export default App;
