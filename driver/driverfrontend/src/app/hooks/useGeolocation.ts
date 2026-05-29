import { useState, useEffect, useCallback, useRef } from "react";

interface GeoState {
  lat: number;
  lng: number;
  address: string;
  accuracy: number | null;
  error: string | null;
  loading: boolean;
  source: 'gps' | 'ip' | 'static';
}

const DEFAULT_CENTER: [number, number] = [33.96, 118.3];
const GOOD_ACCURACY = 1000;
const GPS_TIMEOUT = 3000;
const WIFI_TIMEOUT = 5000;

const IP_SERVICES = [
  'https://freegeoip.app/json/',
  'https://ipwho.is/',
  'https://ipinfo.io/json',
];

async function fallbackIpLocation(): Promise<{ lat: number; lng: number; ip: string } | null> {
  for (const url of IP_SERVICES) {
    try {
      const res = await fetch(url, { signal: AbortSignal.timeout(3000) });
      const data = await res.json();
      if (url.includes('freegeoip') && data.latitude && data.longitude) {
        return { lat: data.latitude, lng: data.longitude, ip: data.ip || '' };
      }
      if (url.includes('ipwho') && data.latitude && data.longitude) {
        return { lat: data.latitude, lng: data.longitude, ip: data.ip || '' };
      }
      if (url.includes('ipinfo') && data.loc) {
        const [lat, lng] = data.loc.split(',');
        return { lat: parseFloat(lat), lng: parseFloat(lng), ip: data.ip || '' };
      }
    } catch { continue; }
  }
  return null;
}

function getCurrentPosition(options: PositionOptions): Promise<GeolocationPosition> {
  return new Promise((resolve, reject) => {
    navigator.geolocation.getCurrentPosition(resolve, reject, options);
  });
}

async function tryGetLocation(signal: AbortSignal): Promise<{ lat: number; lng: number; accuracy: number; source: 'gps' | 'ip' }> {
  if (navigator.geolocation) {
    // Stage 1: GPS (3s timeout)
    try {
      const pos = await getCurrentPosition({ enableHighAccuracy: true, timeout: GPS_TIMEOUT, maximumAge: 0 });
      if (pos.coords.accuracy < GOOD_ACCURACY) {
        return { lat: pos.coords.latitude, lng: pos.coords.longitude, accuracy: pos.coords.accuracy, source: 'gps' };
      }
      // GPS signal weak, continue to stage 2
    } catch {
      // GPS timeout, continue to stage 2
    }

    // Stage 2: WiFi (5s timeout)
    try {
      const pos = await getCurrentPosition({ enableHighAccuracy: true, timeout: WIFI_TIMEOUT, maximumAge: 0 });
      return { lat: pos.coords.latitude, lng: pos.coords.longitude, accuracy: pos.coords.accuracy, source: 'gps' };
    } catch {
      // WiFi also failed, continue to IP fallback
    }
  }

  // Stage 3: IP fallback
  if (signal.aborted) throw new DOMException("Aborted", "AbortError");
  const ip = await fallbackIpLocation();
  if (ip) {
    return { lat: ip.lat, lng: ip.lng, accuracy: 5000, source: 'ip' };
  }
  throw new Error("All location methods failed");
}

export function useGeolocation() {
  const [state, setState] = useState<GeoState>({
    lat: DEFAULT_CENTER[0], lng: DEFAULT_CENTER[1], address: "", accuracy: null, error: null, loading: true, source: 'static',
  });
  const abortRef = useRef<AbortController | null>(null);

  const start = useCallback(async () => {
    abortRef.current?.abort();
    const controller = new AbortController();
    abortRef.current = controller;

    setState((s) => ({ ...s, loading: true, error: null }));

    try {
      const result = await tryGetLocation(controller.signal);
      if (!controller.signal.aborted) {
        console.log(`[定位] ${result.source === 'gps' ? 'GPS/WiFi' : 'IP定位'}: ${result.lat.toFixed(4)}, ${result.lng.toFixed(4)} (精度: ±${Math.round(result.accuracy)}m)`);
        setState({ lat: result.lat, lng: result.lng, address: "", accuracy: result.accuracy, error: null, loading: false, source: result.source });
      }
    } catch (e: any) {
      if (e.name === "AbortError") return;
      if (!controller.signal.aborted) {
        console.warn(`[定位] 全部失败: ${e.message}`);
        setState({ lat: DEFAULT_CENTER[0], lng: DEFAULT_CENTER[1], address: "", accuracy: null, error: "无法获取位置", loading: false, source: 'static' });
      }
    }
  }, []);

  useEffect(() => {
    start();
    return () => { abortRef.current?.abort(); };
  }, [start]);

  return { ...state, refresh: start };
}

export function useGeolocationWithInterval(refreshInterval?: number) {
  const geo = useGeolocation();
  useEffect(() => {
    if (!refreshInterval || refreshInterval <= 0) return;
    const timer = setInterval(() => geo.refresh(), refreshInterval);
    return () => clearInterval(timer);
  }, [geo.refresh, refreshInterval]);
  return geo;
}
