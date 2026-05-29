import { ReactNode } from "react";
import { Signal, Wifi, BatteryFull } from "lucide-react";

export function PhoneFrame({ children }: { children: ReactNode }) {
  return (
    <div className="mx-auto w-[390px] h-[780px] bg-black rounded-[44px] p-3 shadow-2xl">
      <div className="relative w-full h-full bg-white rounded-[34px] overflow-hidden flex flex-col">
        <div className="flex justify-between items-center px-6 pt-2 pb-1 text-xs text-black z-20">
          <span>9:41</span>
          <div className="absolute left-1/2 -translate-x-1/2 top-1 w-24 h-5 bg-black rounded-full" />
          <div className="flex items-center gap-1">
            <Signal className="w-3 h-3" />
            <Wifi className="w-3 h-3" />
            <BatteryFull className="w-4 h-4" />
          </div>
        </div>
        <div className="flex-1 relative overflow-hidden">{children}</div>
      </div>
    </div>
  );
}