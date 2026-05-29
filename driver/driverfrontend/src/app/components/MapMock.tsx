import { MapPin, Navigation } from "lucide-react";

interface Props {
  from?: string;
  to?: string;
  showCar?: boolean;
  className?: string;
}

export function MapMock({ from, to, showCar, className = "" }: Props) {
  return (
    <div className={`relative overflow-hidden bg-gradient-to-br from-emerald-50 via-sky-50 to-indigo-50 ${className}`}>
      <svg className="absolute inset-0 w-full h-full" preserveAspectRatio="none">
        <defs>
          <pattern id="gridp" width="40" height="40" patternUnits="userSpaceOnUse">
            <path d="M 40 0 L 0 0 0 40" fill="none" stroke="#cbd5e1" strokeWidth="0.5" />
          </pattern>
        </defs>
        <rect width="100%" height="100%" fill="url(#gridp)" />
        <path d="M 0 120 Q 150 100 300 180 T 600 150" stroke="#fbbf24" strokeWidth="6" fill="none" opacity="0.6" />
        <path d="M 80 0 Q 120 150 200 300 T 280 600" stroke="#fbbf24" strokeWidth="6" fill="none" opacity="0.6" />
        <path d="M 0 260 L 600 300" stroke="#fde68a" strokeWidth="8" fill="none" opacity="0.5" />
        {(from || to) && (
          <path d="M 70 80 Q 180 140 230 230 T 380 360" stroke="#f97316" strokeWidth="4" fill="none" strokeDasharray="8 4" />
        )}
      </svg>

      {from && (
        <div className="absolute top-[12%] left-[12%] flex flex-col items-center">
          <div className="bg-emerald-500 text-white text-[10px] px-2 py-0.5 rounded-full whitespace-nowrap shadow">起 · {from}</div>
          <div className="w-3 h-3 rounded-full bg-emerald-500 border-2 border-white shadow mt-1" />
        </div>
      )}
      {to && (
        <div className="absolute bottom-[22%] right-[14%] flex flex-col items-center">
          <div className="bg-rose-500 text-white text-[10px] px-2 py-0.5 rounded-full whitespace-nowrap shadow flex items-center gap-1"><MapPin className="w-3 h-3" />{to}</div>
          <div className="w-3 h-3 rounded-full bg-rose-500 border-2 border-white shadow mt-1" />
        </div>
      )}

      {showCar && (
        <div className="absolute top-[45%] left-[40%]">
          <div className="bg-white rounded-full p-2 shadow-lg border-2 border-orange-400 animate-pulse">
            <Navigation className="w-4 h-4 text-orange-500 rotate-45" />
          </div>
        </div>
      )}

      <div className="absolute bottom-3 right-3 w-10 h-10 bg-white rounded-full shadow flex items-center justify-center">
        <Navigation className="w-4 h-4 text-gray-600" />
      </div>
    </div>
  );
}