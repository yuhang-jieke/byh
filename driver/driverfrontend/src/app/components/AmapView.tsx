import { useRef, useEffect, useState } from "react";
import { loadAMap, reverseGeocode, searchDrivingRoute } from "../utils/amap";
import { useGeolocationWithInterval } from "../hooks/useGeolocation";

interface Props {
  from?: string;
  to?: string;
  fromLat?: number;
  fromLng?: number;
  toLat?: number;
  toLng?: number;
  showCar?: boolean;
  showToMarker?: boolean;
  centerOnDriver?: boolean;
  refreshInterval?: number;
  onLocationChange?: (lat: number, lng: number) => void;
  suppressOffsetPrompt?: boolean;
  className?: string;
  driverLat?: number;
  driverLng?: number;
  showDrivingRoute?: boolean;
  routeFromLat?: number;
  routeFromLng?: number;
  routeToLat?: number;
  routeToLng?: number;
}

// 计算地球上两点距离 (km)
function getDistance(lat1: number, lon1: number, lat2: number, lon2: number): number {
  const R = 6371;
  const dLat = (lat2 - lat1) * Math.PI / 180;
  const dLon = (lon2 - lon1) * Math.PI / 180;
  const a = Math.sin(dLat / 2) ** 2 + Math.cos(lat1 * Math.PI / 180) * Math.cos(lat2 * Math.PI / 180) * Math.sin(dLon / 2) ** 2;
  return R * 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));
}

export function AmapView({ from, to, fromLat = 0, fromLng = 0, toLat = 0, toLng = 0, showCar, showToMarker = true, centerOnDriver, refreshInterval, onLocationChange, suppressOffsetPrompt, className = "", driverLat = 0, driverLng = 0, showDrivingRoute, routeFromLat, routeFromLng, routeToLat, routeToLng }: Props) {
  const containerRef = useRef<HTMLDivElement>(null);
  const mapRef = useRef<any>(null);
  const overlaysRef = useRef<any[]>([]);
  const locDot = useRef<any>(null);
  const locRing = useRef<any>(null);
  const carMarker = useRef<any>(null);
  const routeSet = useRef(false);
  const coordsRef = useRef<{ lat: number, lng: number, refresh: () => void }>({ lat: 0, lng: 0, refresh: () => {} });
  
  const [reversedAddress, setReversedAddress] = useState("");
  const [mapReady, setMapReady] = useState(false);
  
  // 弹窗状态
  const [showDistPrompt, setShowDistPrompt] = useState(false);
  const [promptPt, setPromptPt] = useState<[number, number] | null>(null);
  const [promptDist, setPromptDist] = useState(0);
  const hasPromptedRef = useRef(false);

  const { lat, lng, accuracy, error, loading, refresh, source } = useGeolocationWithInterval(refreshInterval);

  // 更新坐标
  useEffect(() => {
    coordsRef.current = { lat, lng, refresh };
  }, [lat, lng, refresh]);

  // 上报位置变更
  useEffect(() => {
    if (lat && lng && onLocationChange) {
      onLocationChange(lat, lng);
    }
  }, [lat, lng, onLocationChange]);

  // 1. 初始化地图
  useEffect(() => {
    if (!containerRef.current || mapRef.current) return;
    loadAMap().then((AMap) => {
      if (!containerRef.current) return;
      mapRef.current = new AMap.Map(containerRef.current, {
        zoom: 14, center: [118.3, 33.95], viewMode: "2D",
      });
      setMapReady(true);
    });
  }, []);

  // 2. 距离检测与弹窗逻辑
  useEffect(() => {
    if (suppressOffsetPrompt) return;
    const map = mapRef.current;
    if (!map || !lat || !lng || hasPromptedRef.current) return;

    (async () => {
      const center = map.getCenter();
      const centerLat = center.getLat();
      const centerLng = center.getLng();
      const dist = getDistance(centerLat, centerLng, lat, lng);

      // 超过 1 公里且未弹出过提示，则显示切换弹窗
      if (dist > 1.0) {
        setPromptPt([lng, lat]);
        setPromptDist(dist);
        // 如果已经解析出地址，直接更新弹窗显示内容
        const addr = await reverseGeocode(lat, lng);
        setReversedAddress(addr);
        setShowDistPrompt(true);
      }
    })();
  }, [lat, lng, mapReady]);

  // 3. 处理切换确认
  const handleConfirmSwitch = () => {
    const map = mapRef.current;
    if (map && promptPt) {
      map.setCenter(promptPt, true, Math.max(map.getZoom(), 16));
      hasPromptedRef.current = true;
    }
    setShowDistPrompt(false);
  };

  const handleCancelSwitch = () => {
    setShowDistPrompt(false);
    hasPromptedRef.current = true;
  };

  // 4. 更新定位点 + 司机车标（轮询坐标更新时仅更新车标位置）
  useEffect(() => {
    const map = mapRef.current;
    if (!map) return;

    // 优先使用传入的司机坐标（接驾阶段轮询）
    const posLat = driverLat || lat;
    const posLng = driverLng || lng;
    if (!posLat || !posLng) return;

    (async () => {
      try {
        const AMap = await loadAMap();
        const pt = new AMap.LngLat(posLng, posLat);

        // 更新定位圆点
        if (locDot.current) map.remove(locDot.current);
        if (locRing.current) map.remove(locRing.current);

        locRing.current = new AMap.Circle({
          center: pt,
          radius: 28,
          strokeColor: "#93c5fd",
          strokeWeight: 2,
          strokeOpacity: 0.4,
          fillColor: "#93c5fd",
          fillOpacity: 0.2,
          zIndex: 10
        });
        locDot.current = new AMap.CircleMarker({
          center: pt,
          radius: 6,
          strokeColor: "#fff",
          strokeWeight: 2,
          fillColor: "#3b82f6",
          fillOpacity: 1,
          zIndex: 40
        });
        map.add([locRing.current, locDot.current]);

        // 更新车标位置
        if (carMarker.current) {
          carMarker.current.setPosition(pt);
        }

        if (centerOnDriver && !showDistPrompt) {
          map.setCenter(pt, true, 16);
        }

        if (!reversedAddress) {
          reverseGeocode(posLat, posLng).then(setReversedAddress).catch(() => {});
        }
      } catch {}
    })();
  }, [lat, lng, driverLat, driverLng, centerOnDriver]);

  // 5. 绘制路线（仅一次，避免高频请求高德接口）
  useEffect(() => {
    const map = mapRef.current;
    if (!map || !mapReady) return;

    // 当 from 变化时，重置 routeSet 标志
    if (routeSet.current && from) {
      // 已经绘制过，跳过
      return;
    }
    if (!from) {
      // 没有 from 值，跳过
      return;
    }

    // 有地址文本就继续绘制（坐标可能在原型数据中缺失）
    const hasFrom = from;
    const hasTo = to;

    overlaysRef.current.forEach((o) => map.remove(o));
    overlaysRef.current = [];
    routeSet.current = true;

    (async () => {
      try {
        const AMap = await loadAMap();
        console.log(`[路线绘制] from="${from}", to="${to}"`);

        // 优先使用后端传入的经纬度坐标
        let srcPt: any;
        if (fromLat !== 0 && fromLng !== 0) {
          srcPt = new AMap.LngLat(fromLng, fromLat);
        } else if (centerOnDriver && coordsRef.current.lat && coordsRef.current.lng) {
          // 接单首页：使用 GPS 定位作为起点
          const c = coordsRef.current;
          srcPt = new AMap.LngLat(c.lng, c.lat);
        } else {
          console.warn(`[路线绘制] 缺少起点坐标，跳过绘制`);
          return;
        }

        const srcMark = new AMap.Marker({
          position: srcPt,
          content: `<div style="display:flex;align-items:center;gap:4px;background:white;padding:3px 8px;border-radius:6px;box-shadow:0 2px 6px rgba(0,0,0,0.15);font-size:12px;white-space:nowrap;"><div style="width:8px;height:8px;border-radius:50%;background:#10b981;"></div><b>${from}</b></div>`,
          offset: new AMap.Pixel(-15, -15),
          zIndex: 30
        });
        overlaysRef.current.push(srcMark); map.add(srcMark);

        if (to && hasTo) {
          let dstPt: any;
          if (toLat !== 0 && toLng !== 0) {
            dstPt = new AMap.LngLat(toLng, toLat);
          } else {
            console.warn(`[路线绘制] 缺少终点坐标，跳过终点标记`);
          }

          if (dstPt) {
            // 根据 showToMarker 控制是否显示目的地标记
            if (showToMarker) {
              const dstMark = new AMap.Marker({
                position: dstPt,
                content: `<div style="display:flex;align-items:center;gap:4px;background:white;padding:3px 8px;border-radius:6px;box-shadow:0 2px 6px rgba(0,0,0,0.15);font-size:12px;white-space:nowrap;"><div style="width:8px;height:8px;border-radius:50%;background:#f43f5e;"></div><b>${to}</b></div>`,
                offset: new AMap.Pixel(-15, -15),
                zIndex: 30
              });
              overlaysRef.current.push(dstMark); map.add(dstMark);
            }

            const srcArr: [number, number] = [fromLat !== 0 ? fromLat : srcPt.getLat(), fromLng !== 0 ? fromLng : srcPt.getLng()];
            const dstArr: [number, number] = [toLat !== 0 ? toLat : dstPt.getLat(), toLng !== 0 ? toLng : dstPt.getLng()];
            if (srcArr[0] && dstArr[0]) {
              console.log(`[路线绘制] 调用路线规划：src=${srcArr}, dst=${dstArr}`);
              const route = await searchDrivingRoute(srcArr, dstArr);
              console.log(`[路线绘制] 路线结果:`, route);
              if (route.paths?.length > 0) {
                const path = route.paths[0].steps.map((s: any) => s.polyline).flat();
                console.log(`[路线绘制] 路径点数量：${path?.length}`);
                if (path && path.length > 0) {
                  const line = new AMap.Polyline({
                    path,
                    strokeColor: "#FF6600",
                    strokeWeight: 5,
                    strokeOpacity: 0.85,
                    zIndex: 20
                  });
                  overlaysRef.current.push(line); map.add(line);
                } else {
                  console.warn('[路线绘制] 路径为空，跳过绘制');
                }
              } else {
                console.warn('[路线绘制] 无路线数据');
              }
            }
            map.setFitView(overlaysRef.current, false, [80, 80, 80, 300]);
          }
        } else {
          // 只有起点（接驾阶段）
          map.setCenter(srcPt, true, 14);
          if (showCar) {
            const car = new AMap.Marker({
              position: srcPt,
              content: `<div style="font-size:24px;filter:drop-shadow(0 2px 3px rgba(0,0,0,0.3));">🚗</div>`,
              offset: new AMap.Pixel(0, 0),
              zIndex: 35
            });
            carMarker.current = car;
            overlaysRef.current.push(car); map.add(car);
          }
        }
      } catch (e) { console.error(`[路线绘制] 异常:`, e); }
    })();
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [from, to, fromLat, fromLng, toLat, toLng, showCar, showToMarker, mapReady]);

  // 6. 绘制接驾导航路线（司机当前位置 → 上车点）
  const drivingRouteRef = useRef<any>(null);
  const drivingRef = useRef<any>(null);
  useEffect(() => {
    if (!showDrivingRoute || !mapRef.current || !mapReady) return;
    if (!routeFromLat || !routeFromLng || !routeToLat || !routeToLng) return;

    const map = mapRef.current;

    // 验证坐标有效性
    const isValidCoord = (lat: number, lng: number) => {
      return lat !== 0 && lng !== 0 && !isNaN(lat) && !isNaN(lng) &&
             lat >= -90 && lat <= 90 && lng >= -180 && lng <= 180;
    };

    if (!isValidCoord(routeFromLat, routeFromLng) || !isValidCoord(routeToLat, routeToLng)) {
      console.warn(`[接驾路线] 坐标无效：from(${routeFromLat},${routeFromLng}), to(${routeToLat},${routeToLng})`);
      return;
    }

    (async () => {
      try {
        const AMap = await loadAMap();

        // 清除旧的路线
        if (drivingRouteRef.current) {
          map.remove(drivingRouteRef.current);
          drivingRouteRef.current = null;
        }

        const fromPt = new AMap.LngLat(routeFromLng, routeFromLat);
        const toPt = new AMap.LngLat(routeToLng, routeToLat);

        console.log(`[接驾路线] 规划路线：from(${routeFromLat.toFixed(4)},${routeFromLng.toFixed(4)}) → to(${routeToLat.toFixed(4)},${routeToLng.toFixed(4)})`);

        // 如果司机已经在上车点附近（<20米），跳过路线规划
        const dist = getDistance(routeFromLat, routeFromLng, routeToLat, routeToLng);
        if (dist < 0.02) {
          console.log(`[接驾路线] 司机已在上车点附近（${dist.toFixed(3)}km），跳过路线规划`);
          return;
        }

        // 尝试 JS SDK 规划路线，失败则画直线
        let path: any[];
        let distance = 0;
        let duration = 0;
        try {
          const { planDrivingRoute } = await import('../utils/amap');
          const resp = await planDrivingRoute(routeFromLat, routeFromLng, routeToLat, routeToLng);
          if (resp && resp.polyline && resp.polyline.length > 0) {
            path = resp.polyline.map((p: { lng: number; lat: number }) => [p.lng, p.lat]);
            distance = resp.distance;
            duration = resp.duration;
            console.log(`[接驾路线] 规划成功，距离：${resp.distance}米，预计：${resp.duration}秒`);
          } else {
            throw new Error('empty route');
          }
        } catch (e) {
          console.warn(`[接驾路线] SDK 规划失败，使用直线:`, e);
          path = [[routeFromLng, routeFromLat], [routeToLng, routeToLat]];
          distance = Math.round(dist * 1000);
        }

        // 绘制路线（SDK 规划的路线 或 直线）
        const polyline = new AMap.Polyline({
          path,
          strokeColor: "#FF6600",
          strokeWeight: 5,
          strokeOpacity: 0.85,
          zIndex: 20
        });

        map.add(polyline);
        drivingRouteRef.current = polyline;

        console.log(`[接驾路线] 规划完成，距离：${distance}米，预计：${duration}秒`);

      } catch (e) {
        console.error(`[接驾路线] 异常:`, e);
      }
    })();

    return () => {
      if (drivingRouteRef.current && map) {
        map.remove(drivingRouteRef.current);
      }
      if (drivingRef.current) {
        drivingRef.current = null;
      }
    };
  }, [showDrivingRoute, routeFromLat, routeFromLng, routeToLat, routeToLng, mapReady]);

  // 7. 定位按钮
  useEffect(() => {
    if (!containerRef.current) return;
    const btn = document.createElement("button");
    btn.innerHTML = "📍";
    btn.style.cssText = "width:36px;height:36px;background:white;border-radius:50%;box-shadow:0 2px 10px rgba(0,0,0,0.2);border:none;cursor:pointer;font-size:20px;display:flex;align-items:center;justify-content:center;position:absolute;bottom:12px;right:12px;z-index:1000;transition:transform 0.2s;";
    btn.onmouseenter = () => btn.style.transform = "scale(1.1)";
    btn.onmouseleave = () => btn.style.transform = "scale(1)";
    
    btn.onclick = () => {
      const c = coordsRef.current;
      const map = mapRef.current;
      const btnLat = driverLat || c.lat;
      const btnLng = driverLng || c.lng;
      if (map && btnLat && btnLng) {
        map.setCenter([btnLng, btnLat], true, 16);
        hasPromptedRef.current = true;
        setShowDistPrompt(false);
      } else { refresh(); }
    };
    containerRef.current.appendChild(btn);
  }, []);

  // 状态 UI
  const status = error ? "error" : loading ? "loading" : "ok";
  const bg = status === "error" ? "bg-amber-50/95 text-amber-700" : status === "loading" ? "bg-blue-50/95 text-blue-700" : "bg-emerald-50/95 text-emerald-700";
  const srcLabel = source === 'ip' ? '网络IP' : "GPS";
  const displayLat = driverLat || lat;
  const displayLng = driverLng || lng;
  const txt = status === "error" ? error : status === "loading" ? `${srcLabel} 定位中...` : (reversedAddress || `${displayLat.toFixed(4)}, ${displayLng.toFixed(4)}`);

  return (
    <div ref={containerRef} className={`relative overflow-hidden ${className}`} style={{ position: "relative", zIndex: 0, width: "100%", height: "100%" }}>
      {/* 定位状态条 */}
      <div className={`absolute top-2 left-2 rounded-full px-3 py-1.5 text-[11px] shadow-md flex items-center gap-1.5 max-w-[260px] z-[1001] ${bg}`}>
        {status === "loading" && <div className="w-2 h-2 border-2 border-current border-t-transparent rounded-full animate-spin shrink-0" />}
        {!loading && error && <span className="shrink-0">⚠️</span>}
        {source === 'ip' && <span className="shrink-0 text-[9px] bg-blue-100 text-blue-600 px-1 rounded">网络</span>}
        <span className="truncate font-medium">{txt}</span>
        {status !== "error" && accuracy !== null && <span className="shrink-0 opacity-60">±{Math.round(accuracy)}m</span>}
      </div>

      {error && (
        <button onClick={refresh} className="absolute top-2 right-2 bg-blue-500 hover:bg-blue-600 text-white text-[11px] px-3 py-1.5 rounded-full shadow z-[1001] flex items-center gap-1 transition-colors">🔄 重试</button>
      )}

      {/* 距离偏离切换弹窗 */}
      {showDistPrompt && promptPt && (
        <div className="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 bg-white/95 backdrop-blur-sm rounded-xl shadow-2xl border border-gray-100 p-4 w-[280px] z-[9999] flex flex-col gap-3">
          <div className="flex items-start gap-3">
            <div className="w-10 h-10 bg-blue-50 rounded-full flex items-center justify-center text-xl shrink-0">🚀</div>
            <div>
              <h3 className="font-bold text-gray-900 text-sm">检测到您的位置偏移</h3>
              <p className="text-xs text-gray-500 mt-0.5 leading-relaxed">
                当前位置 ({reversedAddress || "已定位"}) <br/>
                距离地图中心约 <span className="font-bold text-blue-600">{promptDist.toFixed(1)} 公里</span>，是否切换？
              </p>
            </div>
          </div>
          <div className="grid grid-cols-2 gap-2">
            <button onClick={handleCancelSwitch} className="bg-gray-100 hover:bg-gray-200 text-gray-700 text-xs font-medium py-2.5 rounded-lg transition-colors">留在当前</button>
            <button onClick={handleConfirmSwitch} className="bg-blue-500 hover:bg-blue-600 text-white text-xs font-medium py-2.5 rounded-lg shadow transition-colors">切换到定位</button>
          </div>
        </div>
      )}
    </div>
  );
}