import { createRoot } from "react-dom/client";
import { StoreProvider } from "./app/store";
import { DriverApp } from "./app/components/DriverApp";
import "./styles/index.css";

function App() {
  return (
    <StoreProvider>
      <div className="min-h-screen w-full bg-gradient-to-br from-slate-100 via-gray-50 to-zinc-100">
        <div className="max-w-screen-2xl mx-auto p-5">
          <div className="flex justify-center py-4">
            <div className="flex flex-col items-center">
              <div className="flex items-center gap-1.5 text-xs text-gray-500 mb-3">
                <div className="w-2 h-2 rounded-full bg-emerald-500" />
                司机端 · 花小猪打车
              </div>
              <DriverApp />
            </div>
          </div>
        </div>
      </div>
    </StoreProvider>
  );
}

createRoot(document.getElementById("root")!).render(<App />);