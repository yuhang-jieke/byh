import { useState, useEffect, useRef } from "react";
import { PhoneFrame } from "./PhoneFrame";
import { AmapView } from "./AmapView";
import { useStore, Order, statusLabel, statusColor } from "../store";
import {
  Power, Home, ClipboardList, User, Star, TrendingUp, DollarSign,
  ChevronRight, Phone, Navigation, Shield, Award, Bell, Settings,
  MessageCircle, ChevronLeft, Zap, CheckCircle2, AlertTriangle, Car,
  Gift, Headphones, X, MapPin, Clock, CreditCard, AlertCircle,
  PhoneCall, FileText, Lock, LogOut, Activity, Volume2, WifiOff,
  Wifi, Copy, ChevronDown
} from "lucide-react";
import { DigitalHumanBubble, DigitalHumanPanel } from "./DigitalHuman";
import { createSpeechEngine, isSpeechSupported, SpeechState } from "../utils/speechEngine";

type Tab = "home" | "orders" | "me";
type MeStage = "main" | "income" | "service" | "car" | "sos" | "settings";

/* ============================================================
   滑动确认组件
   ============================================================ */
function SlideToConfirm({
  label, onConfirm,
  gradient = "from-emerald-500 to-teal-500",
  emoji = "🚀", resetKey,
}: {
  label: string; onConfirm: () => void;
  gradient?: string; emoji?: string; resetKey?: string | number;
}) {
  const [offset, setOffset] = useState(0);
  const [confirmed, setConfirmed] = useState(false);
  const trackRef = useRef<HTMLDivElement>(null);
  const dragging = useRef(false);
  const startX = useRef(0);
  const startOffset = useRef(0);

  useEffect(() => { setOffset(0); setConfirmed(false); }, [resetKey]);

  const getMax = () => (trackRef.current ? trackRef.current.offsetWidth - 52 : 200);
  const onStart = (clientX: number) => {
    if (confirmed) return;
    dragging.current = true; startX.current = clientX; startOffset.current = offset;
  };
  const onMove = (clientX: number) => {
    if (!dragging.current) return;
    setOffset(Math.max(0, Math.min(clientX - startX.current + startOffset.current, getMax())));
  };
  const onEnd = () => {
    if (!dragging.current) return;
    dragging.current = false;
    if (offset >= getMax() * 0.82) { setOffset(getMax()); setConfirmed(true); setTimeout(onConfirm, 350); }
    else setOffset(0);
  };
  const pct = trackRef.current ? Math.min(100, (offset / getMax()) * 100) : 0;

  return (
    <div ref={trackRef} className="relative rounded-full overflow-hidden select-none"
      style={{ background: "#f0fdf4", height: 52, touchAction: "none" }}
      onMouseMove={e => onMove(e.clientX)} onMouseUp={onEnd} onMouseLeave={onEnd}
      onTouchMove={e => onMove(e.touches[0].clientX)} onTouchEnd={onEnd}>
      <div className={`absolute left-0 top-0 h-full bg-gradient-to-r ${gradient} rounded-full`}
        style={{ width: `${6 + pct * 0.88}%`, opacity: 0.25, transition: dragging.current ? "none" : "width 0.3s ease" }} />
      <div className="absolute inset-0 flex items-center justify-center text-xs text-gray-400 pointer-events-none select-none">
        {confirmed
          ? <span className="text-emerald-600 font-medium">✓ 已确认</span>
          : <><span className="mr-14">{label}</span><span className="text-gray-300">›› 向右滑动</span></>}
      </div>
      <div className={`absolute top-1 left-1 w-11 h-11 bg-gradient-to-r ${gradient} rounded-full shadow-md flex items-center justify-center text-white text-base font-bold cursor-grab active:cursor-grabbing z-10`}
        style={{ transform: `translateX(${offset}px)`, transition: dragging.current ? "none" : "transform 0.3s cubic-bezier(.4,0,.2,1)" }}
        onMouseDown={e => { e.preventDefault(); onStart(e.clientX); }}
        onTouchStart={e => onStart(e.touches[0].clientX)}>
        {confirmed ? "✓" : emoji}
      </div>
    </div>
  );
}

/* ============================================================
   主 App
   ============================================================ */
export function DriverApp() {
  const { orders, driverInfo, poolOrders, acceptOrder, rejectOrder, arriveOrder, startTrip, endTrip, verifyPassenger, cancelOrder, toggleDriverStatus, setDriverListening, reportLocation, loadOrders, loadDriverInfo, loadPoolOrders, grabOrder, getOrderDetail, showToast: storeToast, startLocationUpdates, stopLocationUpdates } = useStore();
  const [tab, setTab] = useState<Tab>("home");
  const [meStage, setMeStage] = useState<MeStage>("main");
  const [toast, setToast] = useState<string | null>(null);
  const [selectedHistoryId, setSelectedHistoryId] = useState<string | null>(null);
  const [selectedOrder, setSelectedOrder] = useState<Order | null>(null);
  const [cancellationAlert, setCancellationAlert] = useState<{ name: string } | null>(null);
  const [withdrawalOpen, setWithdrawalOpen] = useState(false);
  const [sosOpen, setSosOpen] = useState(false);
  const [notifOpen, setNotifOpen] = useState(false);
  const [dispatchModal, setDispatchModal] = useState<Order | null>(null);
  const [dhOpen, setDhOpen] = useState(false);
  const { voiceMode, setVoiceMode } = useStore();
  const [voiceState, setVoiceState] = useState<SpeechState>("idle");
  const voiceEngineRef = useRef<ReturnType<typeof createSpeechEngine> | null>(null);
  // 数字人气泡拖拽
  const bubbleDrag = useRef({ x: 0, y: 0, dragging: false, startX: 0, startY: 0, offsetX: 0, offsetY: 0 });
  const bubbleDragged = useRef(false);

  useEffect(() => {
    if (!voiceMode || !isSpeechSupported()) return;
    let cancelled = false;
    fetch("/api/v1/ai/tts?text=" + encodeURIComponent("请讲"))
      .then(r => r.blob())
      .then(blob => { if (cancelled) return null; const a = new Audio(URL.createObjectURL(blob)); return new Promise(r => { a.onended = r; a.play().catch(r); }); })
      .catch(() => {})
      .then(() => {
        if (cancelled) return;
        const engine = createSpeechEngine({
          onResult: (text) => {
            voiceEngineRef.current?.stopListening();
            handleVoiceCommand(text);
          },
          onError: () => {},
          onStateChange: (s) => setVoiceState(s),
          onIdle: () => {
            fetch("/api/v1/ai/tts?text=" + encodeURIComponent("没什么事我就先退了，有事可以再叫我"))
              .then(r => r.blob()).then(blob => { const a = new Audio(URL.createObjectURL(blob)); a.play(); }).catch(() => {});
            voiceEngineRef.current?.destroy();
            voiceEngineRef.current = null;
            setVoiceState("idle");
            setVoiceMode(false);
          },
        });
        engine.startContinuousListening();
        voiceEngineRef.current = engine;
      });
    return () => {
      cancelled = true;
      voiceEngineRef.current?.destroy();
      voiceEngineRef.current = null;
      setVoiceState("idle");
    };
  }, [voiceMode]);

  function handleVoiceCommand(text: string) {
    fetch("/api/v1/ai/voice-chat", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ message: text, driver_id: 200000001 }),
    })
      .then(r => r.blob())
      .then(blob => {
        const engine = voiceEngineRef.current;
        if (!engine) return;
        const url = URL.createObjectURL(blob);
        const audio = new Audio(url);
        audio.onended = () => {
          URL.revokeObjectURL(url);
          engine.startContinuousListening();
        };
        audio.play();
      })
      .catch(() => {
        voiceEngineRef.current?.startContinuousListening();
      });
  }

  useEffect(() => {
    if (!voiceMode) {
      voiceEngineRef.current?.destroy();
      voiceEngineRef.current = null;
      setVoiceState("idle");
    }
  }, [voiceMode]);

  useEffect(() => {
    if (voiceMode) setToast("🎤 语音助手已开启，可直接说话");
  }, [voiceMode]);

  const driver = driverInfo || {
    id: "200000001", name: "司机", phone: "", plate: "-", car: "-",
    rating: 80, online: false, listening: false, totalOrders: 0, todayEarnings: 0, status: "offline" as const,
  };
  const pendingOrders = orders.filter(o => o.status === "pending");
  const dispatchedOrder = orders.find(o => o.status === "dispatched");
  const myActive = orders.find(o => ["accepted", "arrived", "ongoing"].includes(o.status));
  const myHistory = orders.filter(o => ["completed", "cancelled"].includes(o.status));

  // 检测乘客取消已接订单
  const prevActiveId = useRef<string | null>(null);
  useEffect(() => {
    if (prevActiveId.current && !myActive) {
      const prev = orders.find(o => o.id === prevActiveId.current);
      if (prev?.status === "cancelled") setCancellationAlert({ name: prev.passengerName });
    }
    prevActiveId.current = myActive?.id || null;
  }, [myActive?.id]);

  // 初始加载
  useEffect(() => {
    loadDriverInfo();
  }, []);

  // 切到首页 / 订单页时刷新订单列表
  useEffect(() => {
    if (tab === "home" || tab === "orders") loadOrders();
  }, [tab, loadOrders]);

  // 检测平台派单，弹出确认弹窗
  const lastDispatchId = useRef<string | null>(null);
  useEffect(() => {
    if (dispatchedOrder && dispatchedOrder.id !== lastDispatchId.current) {
      lastDispatchId.current = dispatchedOrder.id;
      setDispatchModal(dispatchedOrder);
    }
  }, [dispatchedOrder]);

  useEffect(() => {
    if (!toast) return;
    const t = setTimeout(() => setToast(null), 2000);
    return () => clearTimeout(t);
  }, [toast]);

  // 加载订单详情
  useEffect(() => {
    if (!selectedHistoryId) return;
    setSelectedOrder(null); // 清空缓存，显示加载状态
    (async () => {
      try {
        const detail = await getOrderDetail(Number(selectedHistoryId));
        if (detail) setSelectedOrder(detail);
      } catch { /* ignore */ }
    })();
  }, [selectedHistoryId, getOrderDetail]);

  // 听单状态时启动位置上报
  useEffect(() => {
    if (driverInfo?.status === "busy") {
      startLocationUpdates();
    } else {
      stopLocationUpdates();
    }
  }, [driverInfo?.status, startLocationUpdates, stopLocationUpdates]);

  // 听单时自动轮询抢单池
  useEffect(() => {
    if (driverInfo?.status !== "busy") return;
    loadPoolOrders();
    const t = setInterval(loadPoolOrders, 15000);
    return () => clearInterval(t);
  }, [driverInfo?.status, loadPoolOrders]);

  function showToast(msg: string) { setToast(msg); }
  function handleComplete() {
    if (myActive) {
      const orderId = parseInt(myActive.id, 10);
      endTrip(orderId);
    }
  }

  // 行程详情（全屏覆盖）
  if (selectedHistoryId) {
    const order = selectedOrder || orders.find(o => o.id === selectedHistoryId);
    if (order) return (
      <PhoneFrame>
        <DriverTripDetailView order={order} onBack={() => { setSelectedHistoryId(null); setSelectedOrder(null); }} onToast={showToast} />
      </PhoneFrame>
    );
  }

  return (
    <PhoneFrame>
      <div className="h-full flex flex-col bg-gray-50 relative overflow-hidden">
        {tab === "home" && !myActive && (
          <DriverHome driver={driver} poolOrders={poolOrders}
            onToggle={toggleDriverStatus}
            onGrab={(id: number) => grabOrder(id)}
            onBell={() => setNotifOpen(true)} onToast={showToast} />
        )}
        {tab === "home" && myActive && (
          <DriverActive order={myActive}
            onToast={showToast}
            onSOS={() => setSosOpen(true)} />
        )}
        {tab !== "home" && (
          <div className="flex-1 overflow-y-auto scrollbar-hide">
            {tab === "orders" && (
              <DriverOrders orders={orders}
                onOrderClick={id => setSelectedHistoryId(id)} onToast={showToast} />
            )}
            {tab === "me" && meStage === "main" && (
              <DriverMe driver={driver} onNav={(s: string) => setMeStage(s as MeStage)}
                onWithdraw={() => setWithdrawalOpen(true)} onToast={showToast} />
            )}
            {tab === "me" && meStage === "income" && (
              <DriverIncomeView driver={driver} onBack={() => setMeStage("main")}
                onWithdraw={() => setWithdrawalOpen(true)} onToast={showToast} />
            )}
            {tab === "me" && meStage === "service" && (
              <DriverServiceView driver={driver} onBack={() => setMeStage("main")} onToast={showToast} />
            )}
            {tab === "me" && meStage === "car" && (
              <DriverCarView driver={driver} onBack={() => setMeStage("main")} onToast={showToast} />
            )}
            {tab === "me" && meStage === "sos" && (
              <DriverSOSView onBack={() => setMeStage("main")} onToast={showToast} />
            )}
            {tab === "me" && meStage === "settings" && (
              <DriverSettingsView driver={driver} onBack={() => setMeStage("main")}
                onToggleOnline={toggleDriverStatus} onToast={showToast} />
            )}
          </div>
        )}

        {/* 底部Tab */}
        <div className="flex border-t bg-white shrink-0">
          {[{ k: "home", icon: Home, label: "接单" }, { k: "orders", icon: ClipboardList, label: "订单" }, { k: "me", icon: User, label: "我的" }]
            .map(({ k, icon: Icon, label }) => (
              <button key={k} onClick={() => { setTab(k as Tab); if (k === "me") setMeStage("main"); }}
                className={`flex-1 py-2 flex flex-col items-center gap-0.5 relative ${tab === k ? "text-emerald-500" : "text-gray-400"}`}>
                <Icon className="w-5 h-5" />
                <span className="text-[10px]">{label}</span>
                {k === "orders" && tab !== "orders" && myHistory.filter(o => o.status === "completed" && Date.now() - o.createdAt < 3600000).length > 0 && (
                  <span className="absolute top-1 right-4 w-1.5 h-1.5 bg-emerald-500 rounded-full" />
                )}
              </button>
            ))}
        </div>

        {/* 新订单浮动提示（非首页时） */}
        {tab !== "home" && driver.status === 'busy' && pendingOrders.length > 0 && !cancellationAlert && (
          <FloatingNewOrderAlert order={pendingOrders[0]}
            onAccept={() => { acceptOrder(parseInt(pendingOrders[0].id, 10)); setTab("home"); }}
            onView={() => setTab("home")} />
        )}

        {/* 派单确认弹窗 */}
        {dispatchModal && (
          <DispatchConfirmModal
            order={dispatchModal}
            onAccept={() => {
              const id = Number(dispatchModal.id);
              setDispatchModal(null);
              acceptOrder(id);
            }}
            onReject={() => {
              const id = Number(dispatchModal.id);
              setDispatchModal(null);
              rejectOrder(id);
            }}
          />
        )}

        {/* 乘客取消弹窗 */}
        {cancellationAlert && (
          <CancellationModal name={cancellationAlert.name} onClose={() => setCancellationAlert(null)} />
        )}

        {/* SOS弹窗 */}
        {sosOpen && <DriverSOSModal onClose={() => setSosOpen(false)} onToast={showToast} />}

        {/* 提现弹窗 */}
        {withdrawalOpen && (
          <WithdrawalModal balance={driver.todayEarnings}
            onClose={() => setWithdrawalOpen(false)} onToast={showToast} />
        )}

        {/* 通知面板 */}
        {notifOpen && (
          <NotificationsPanel orders={myHistory} onClose={() => setNotifOpen(false)} />
        )}

        {/* 数字人悬浮气泡（可拖拽） */}
        {!dhOpen && tab === "home" && (
          <div
            style={{
              position: "absolute", bottom: "3rem", right: "0.75rem", zIndex: 40,
              touchAction: "none", cursor: "grab",
            }}
            onPointerDown={(e) => {
              const r = bubbleDrag.current;
              r.dragging = true; r.startX = e.clientX; r.startY = e.clientY;
              r.offsetX = r.x; r.offsetY = r.y;
              bubbleDragged.current = false;
              e.currentTarget.setPointerCapture(e.pointerId);
            }}
            onPointerMove={(e) => {
              const r = bubbleDrag.current;
              if (!r.dragging) return;
              const dx = e.clientX - r.startX, dy = e.clientY - r.startY;
              if (Math.abs(dx) > 5 || Math.abs(dy) > 5) bubbleDragged.current = true;
              if (bubbleDragged.current) {
                r.x = Math.max(-250, Math.min(250, r.offsetX + dx));
                r.y = Math.max(-400, Math.min(150, r.offsetY + dy));
                e.currentTarget.style.transform = `translate(${r.x}px,${r.y}px)`;
              }
            }}
            onPointerUp={() => { bubbleDrag.current.dragging = false; }}
          >
        <DigitalHumanBubble
              voiceState={voiceState}
              voiceMode={voiceMode}
              onClick={() => {
                if (bubbleDragged.current) { bubbleDragged.current = false; return; }
                voiceMode ? setVoiceMode(false) : setDhOpen(true);
              }}
            />
          </div>
        )}

        {/* 数字人全屏面板 */}
        {dhOpen && <DigitalHumanPanel onClose={() => setDhOpen(false)} />}

        {/* Toast */}
        {toast && (
          <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 bg-black/75 text-white text-xs px-4 py-2 rounded-lg z-[60]">
            {toast}
          </div>
        )}
      </div>
    </PhoneFrame>
  );
}

/* ============================================================
   接单首页
   ============================================================ */
function DriverHome({ driver, poolOrders, onToggle, onGrab, onBell, onToast }: any) {
  const [pulse, setPulse] = useState(false);
  useEffect(() => {
    if (driver.online && poolOrders.length > 0) { setPulse(true); const t = setTimeout(() => setPulse(false), 600); return () => clearTimeout(t); }
  }, [poolOrders.length, driver.online]);

  return (
    <div className="flex flex-col h-full min-h-0">
      {/* 固定顶部：地图 + 收入卡片 */}
      <div className="shrink-0">
        <div className="relative" style={{ height: 200 }}>
          <div className="absolute inset-0">
            <AmapView showCar={driver.online} centerOnDriver suppressOffsetPrompt />
          </div>

          <div className="absolute top-2 left-3 right-3 bg-white rounded-2xl px-3 py-2.5 shadow-md flex items-center gap-3 z-10">
            <div className="w-10 h-10 rounded-full bg-emerald-100 flex items-center justify-center text-xl">👨‍✈️</div>
            <div className="flex-1">
              <div className="text-sm font-medium">{driver.name}</div>
              <div className="text-[10px] text-gray-500 flex items-center gap-1">
                <Star className="w-3 h-3 text-amber-400 fill-current" />{driver.rating}
                <span className="mx-1">·</span>{driver.plate}
              </div>
            </div>
            <button onClick={onToggle}
              className={`px-3.5 py-1.5 rounded-full text-xs flex items-center gap-1.5 transition-all ${driver.status === 'offline' ? "bg-emerald-500 text-white shadow-sm" : driver.status === 'busy' ? "bg-amber-500 text-white" : "bg-gray-200 text-gray-600"}`}>
              <Power className="w-3 h-3" />
              {driver.status === 'offline' ? "出车" : driver.status === 'busy' ? "收车" : "下线"}
            </button>
          </div>

          <button onClick={onBell} className="absolute top-2 right-3 mt-[60px] w-9 h-9 bg-white rounded-full shadow flex items-center justify-center z-10">
            <Bell className="w-4 h-4 text-gray-600" />
            {poolOrders.length > 0 && driver.online && <span className="absolute top-1 right-1 w-2 h-2 bg-rose-500 rounded-full" />}
          </button>
        </div>

        <div className="mx-4 mt-3 mb-3 bg-gradient-to-r from-emerald-500 to-teal-500 rounded-2xl p-4 text-white">
          <div className="flex items-center justify-between mb-2">
            <div className="text-xs opacity-90">今日收入</div>
            <button onClick={() => onToast("查看收入明细")} className="text-[10px] bg-white/20 px-2 py-0.5 rounded-full">明细 ›</button>
          </div>
          <div className="text-3xl font-light mb-2">¥{driver.todayEarnings.toFixed(2)}</div>
          <div className="grid grid-cols-3 gap-2 text-center text-[11px]">
            {[{ l: "📦 完单", v: driver.status === "busy" ? 9 : 8 }, { l: "⏱ 在线", v: "6.2h" }, { l: "⭐ 好评", v: "98%" }]
              .map(x => (
                <div key={x.l} className="bg-white/15 rounded-xl py-2">
                  <div className="opacity-90 mb-0.5">{x.l}</div>
                  <div className="font-medium">{x.v}</div>
                </div>
              ))}
          </div>
        </div>
      </div>

      {/* 可滚动订单区域 */}
      <div className="flex-1 overflow-y-auto scrollbar-hide min-h-0">
        <div className="px-4 pb-16">
          <div className="flex items-center justify-between mb-2">
            <div className="flex items-center gap-2">
              <span className="text-sm font-medium">附近订单</span>
              {driver.status === 'busy' && poolOrders.length > 0 && (
                <span className={`text-[10px] px-2 py-0.5 rounded-full bg-rose-100 text-rose-600 ${pulse ? "scale-110" : ""} transition-transform`}>
                  {poolOrders.length} 单
                </span>
              )}
            </div>
            <span className="text-[10px] text-gray-400">自动刷新中</span>
          </div>

          <div className="space-y-2">
            {driver.status === 'offline' && (
              <div className="text-center py-8 text-xs text-gray-400">
                <div className="text-3xl mb-2">😴</div>
                <div>已收车，点击"出车"重新出车</div>
              </div>
            )}
            {driver.status === 'idle' && (
              <div className="text-center py-8 text-xs text-gray-400">
                <div className="text-3xl mb-2">😌</div>
                <div>在线空闲中，点击"下线"退出</div>
              </div>
            )}
            {driver.status === 'busy' && poolOrders.length === 0 && (
              <div className="text-center py-8 text-xs text-gray-400">
                <div className="text-3xl mb-2 animate-bounce">🚗</div>
                <div>暂无新订单</div>
              </div>
            )}
            {driver.status === 'busy' && poolOrders.map((o: Order) => (
              <div key={o.id} className="bg-orange-50 border border-orange-200 rounded-2xl p-3">
                <div className="flex justify-between items-center mb-2">
                  <div className="flex items-center gap-2">
                    <span className="text-[10px] px-2 py-0.5 bg-orange-500 text-white rounded-full">{o.carType}</span>
                    <span className="text-[10px] text-orange-500 font-medium">抢单</span>
                    <span className="text-[10px] text-gray-500">{o.distanceKm.toFixed(1)}km</span>
                  </div>
                  <span className="text-orange-600 font-medium">¥{o.price}</span>
                </div>
                <div className="flex gap-2 text-xs mb-0.5 items-start">
                  <div className="w-2 h-2 rounded-full bg-emerald-500 mt-1 flex-shrink-0" />
                  <span className="flex-1 text-gray-700 truncate">{o.from}</span>
                </div>
                <div className="flex gap-2 text-xs mb-1 items-start">
                  <div className="w-2 h-2 rounded-full bg-rose-500 mt-1 flex-shrink-0" />
                  <span className="flex-1 text-gray-700 truncate">{o.to}</span>
                </div>
                <SlideToConfirm label="立即抢单" emoji="🚗" gradient="from-orange-500 to-amber-500"
                  onConfirm={() => onGrab(o.id)} resetKey={o.id} />
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}

/* ============================================================
   进行中订单
   ============================================================ */
function DriverActive({ order, onToast, onSOS }: any) {
  const { arriveOrder, startTrip, endTrip, verifyPassenger, loadOrders, driverInfo } = useStore();
  const [verifying, setVerifying] = useState(false);
  const [phoneInput, setPhoneInput] = useState("");
  const orderId = parseInt(order.id, 10);
  const [driverLat, setDriverLat] = useState<number | undefined>();
  const [driverLng, setDriverLng] = useState<number | undefined>();

  // 实时获取司机当前位置用于接驾路线规划
  useEffect(() => {
    if (!navigator.geolocation) return;
    const watchId = navigator.geolocation.watchPosition(
      (pos) => {
        setDriverLat(pos.coords.latitude);
        setDriverLng(pos.coords.longitude);
      },
      () => {},
      { enableHighAccuracy: true, maximumAge: 5000, timeout: 10000 }
    );
    return () => navigator.geolocation.clearWatch(watchId);
  }, []);

  // 加载最新订单状态
  useEffect(() => {
    const timer = setInterval(() => loadOrders(), 10000);
    return () => clearInterval(timer);
  }, [loadOrders]);

  // 启动系统导航到上车点
  const handleNavigate = () => {
    const address = encodeURIComponent(order.from);
    const url = `https://uri.amap.com/navigation?address=${address}&src=driver&mode=car`;
    window.open(url, '_blank');
    onToast("导航已启动");
  };

  const handleArrive = () => arriveOrder(orderId);
  const handleVerify = async () => {
    const ok = await verifyPassenger(orderId, phoneInput);
    if (ok) { setVerifying(false); setPhoneInput(""); await startTrip(orderId); }
  };
  const handleStartTrip = () => setVerifying(true);
  const handleEndTrip = () => endTrip(orderId);

  const actionMap: any = {
    accepted: { label: "已到达上车点", handler: handleArrive, gradient: "from-indigo-500 to-blue-500", emoji: "📍" },
    arrived: { label: "验证乘客并上车", handler: handleStartTrip, gradient: "from-violet-500 to-purple-500", emoji: "🔐" },
  };
  const action = actionMap[order.status];
  const statusInfo: any = {
    accepted: { title: "前往接驾中", sub: "请按导航路线行驶", color: "text-indigo-600", bg: "bg-indigo-50" },
    arrived: { title: "已到达上车点", sub: "等待乘客上车", color: "text-violet-600", bg: "bg-violet-50" },
    ongoing: { title: "行程进行中", sub: `目的地：${order.to}`, color: "text-emerald-600", bg: "bg-emerald-50" },
  };
  const info = statusInfo[order.status] || statusInfo.accepted;

  return (
    <div className="relative h-full">
      <AmapView
        className="h-[340px]"
        // 接驾中显示上车点，行程中显示完整路线
        from={order.status === "accepted" ? order.from : order.status === "ongoing" || order.status === "completed" ? order.from : undefined}
        to={order.status === "ongoing" || order.status === "completed" ? order.to : undefined}
        fromLat={order.fromLat}
        fromLng={order.fromLng}
        toLat={order.toLat}
        toLng={order.toLng}
        showCar
        centerOnDriver
        suppressOffsetPrompt
        // 接驾中隐藏目的地标记
        showToMarker={order.status !== "accepted"}
        // 接驾中显示司机 → 上车点的导航路线
        showDrivingRoute={order.status === "accepted"}
        routeFromLat={driverLat}
        routeFromLng={driverLng}
        routeToLat={order.fromLat}
        routeToLng={order.fromLng}
      />

      {/* SOS浮动按钮 */}
      <button onClick={onSOS} className="absolute top-3 right-3 z-20 bg-rose-500 text-white text-xs font-bold px-3 py-1.5 rounded-full shadow-lg">
        SOS
      </button>

      <div className="absolute left-0 right-0 bottom-0 bg-white rounded-t-3xl p-4 shadow-2xl">
        <div className={`inline-flex items-center gap-1.5 text-xs px-3 py-1.5 rounded-full ${info.bg} ${info.color} mb-3`}>
          <Zap className="w-3 h-3" />{info.title} · {info.sub}
        </div>

        <div className="mb-3 space-y-1.5">
          <div className="flex gap-2 text-sm items-start">
            <div className="w-2 h-2 rounded-full bg-emerald-500 mt-1.5 flex-shrink-0" />
            <span className="flex-1 text-gray-800 truncate">{order.from}</span>
          </div>
          <div className="flex gap-2 text-sm items-start">
            <div className="w-2 h-2 rounded-full bg-rose-500 mt-1.5 flex-shrink-0" />
            <span className="flex-1 text-gray-800 truncate">{order.to}</span>
          </div>
        </div>

        {/* 乘客信息卡 */}
        <div className="bg-gray-50 rounded-2xl p-3 flex items-center gap-3 mb-3">
          <div className="w-10 h-10 rounded-full bg-white border border-gray-200 flex items-center justify-center text-xl">👤</div>
          <div className="flex-1">
            <div className="text-sm font-medium">{order.passengerName}</div>
            <div className="text-[10px] text-gray-500 mt-0.5">{order.passengerPhone}</div>
            {order.note && <div className="text-[10px] text-orange-500 mt-0.5">备注：{order.note}</div>}
          </div>
          <button onClick={() => onToast("正在联系乘客...")} className="w-9 h-9 bg-white border border-gray-200 rounded-full flex items-center justify-center">
            <MessageCircle className="w-4 h-4 text-gray-600" />
          </button>
          <button onClick={() => onToast("正在拨打...")} className="w-9 h-9 bg-emerald-500 rounded-full flex items-center justify-center shadow-sm">
            <Phone className="w-4 h-4 text-white" />
          </button>
        </div>

        {/* 工具栏 */}
        <div className="grid grid-cols-4 gap-2 mb-3 text-xs">
          {[
            { icon: Navigation, label: "导航", color: "text-blue-500", action: handleNavigate },
            { icon: Shield, label: "安全", color: "text-emerald-500", action: onSOS },
            { icon: Headphones, label: "客服", color: "text-violet-500", action: () => onToast("客服通道") },
            { icon: AlertTriangle, label: "异常", color: "text-amber-500", action: () => onToast("上报异常") },
          ].map(({ icon: Icon, label, color, action }) => (
            <button key={label} onClick={action} className="bg-gray-50 rounded-xl py-2.5 flex flex-col items-center gap-1 border border-gray-100">
              <Icon className={`w-4 h-4 ${color}`} /><span>{label}</span>
            </button>
          ))}
        </div>

        {/* 手机号验证输入 */}
        {verifying && (
          <div className="mb-3">
            <div className="text-xs text-gray-500 mb-1">请输入乘客手机号后4位</div>
            <div className="flex gap-2 mb-2">
              <input
                type="text" maxLength={4} inputMode="numeric"
                value={phoneInput} onChange={e => setPhoneInput(e.target.value.replace(/\D/g, ""))}
                className="flex-1 border border-gray-300 rounded-xl px-4 py-2.5 text-center text-lg tracking-widest focus:border-violet-500 outline-none"
                placeholder="请输入后4位"
              />
            </div>
            <div className="flex gap-2">
              <button onClick={() => { setVerifying(false); setPhoneInput(""); }} className="flex-1 bg-gray-100 text-gray-600 py-3 rounded-full text-sm">取消</button>
              <button onClick={handleVerify} disabled={phoneInput.length !== 4} className="flex-1 bg-violet-500 text-white py-3 rounded-full text-sm disabled:opacity-50">确认验证</button>
            </div>
          </div>
        )}

        {action && !verifying && (
          <SlideToConfirm label={action.label} emoji={action.emoji} gradient={action.gradient}
            onConfirm={action.handler} resetKey={order.status} />
        )}
        {order.status === "ongoing" && !verifying && (
          <SlideToConfirm label={`完成行程 · ¥${order.price}`} emoji="🏁"
            gradient="from-orange-500 to-rose-500" onConfirm={handleEndTrip}
            resetKey={order.id + "-ongoing"} />
        )}
      </div>
    </div>
  );
}

/* ============================================================
   订单历史
   ============================================================ */
function DriverOrders({ orders, onOrderClick, onToast }: { orders: Order[]; onOrderClick: (id: string) => void; onToast: (m: string) => void }) {
  const [tab, setTab] = useState<"today" | "all">("today");
  const { loadOrders } = useStore();
  const allOrders = orders;
  const todayStr = (() => {
    const n = new Date();
    return `${n.getFullYear()}-${String(n.getMonth() + 1).padStart(2, "0")}-${String(n.getDate()).padStart(2, "0")}`;
  })();
  const isToday = (ts: number) => {
    const d = new Date(ts);
    return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}` === todayStr;
  };
  const completed = orders.filter(o => o.status === "completed");
  const today = completed.filter(o => isToday(o.createdAt));
  const todayList = allOrders.filter(o => isToday(o.createdAt));
  const list = tab === "today" ? todayList : allOrders;
  const todayGMV = today.reduce((s, o) => s + o.price, 0);

  const handleTabChange = (k: string) => {
    if (k === "all") {
      loadOrders(undefined, 0, true);
    }
    setTab(k as "today" | "all");
  };

  return (
    <div className="min-h-full bg-gray-50">
      <div className="bg-gradient-to-r from-emerald-500 to-teal-500 px-4 py-5">
        <div className="grid grid-cols-3 gap-3 text-white text-center">
          {[
            { val: today.length, label: "今日完单" },
            { val: `¥${todayGMV.toFixed(0)}`, label: "今日收入" },
            { val: completed.length, label: "历史总单" },
          ].map(c => (
            <div key={c.label} className="bg-white/20 rounded-xl py-3">
              <div className="text-xl font-medium">{c.val}</div>
              <div className="text-[10px] opacity-80 mt-0.5">{c.label}</div>
            </div>
          ))}
        </div>
      </div>

      <div className="flex bg-white border-b">
        {[["today", "今日"], ["all", "全部历史"]].map(([k, l]) => (
          <button key={k} onClick={() => handleTabChange(k)}
            className={`flex-1 py-3 text-sm relative ${tab === k ? "text-emerald-600" : "text-gray-500"}`}>
            {l}
            {tab === k && <div className="absolute bottom-0 left-1/2 -translate-x-1/2 w-8 h-0.5 bg-emerald-500 rounded" />}
          </button>
        ))}
      </div>

      <div className="p-3 space-y-2">
        {list.map(o => (
          <button key={o.id} onClick={() => onOrderClick(o.id)}
            className="w-full bg-white rounded-2xl p-3.5 shadow-sm text-left active:bg-gray-50">
            <div className="flex justify-between mb-2">
              <span className="text-xs text-gray-500">{new Date(o.createdAt).toLocaleString("zh-CN")}</span>
              <div className="flex items-center gap-2">
                {o.rating && (
                  <span className="flex items-center gap-0.5 text-[10px] text-amber-500">
                    <Star className="w-3 h-3 fill-current" />{o.rating}分好评
                  </span>
                )}
                {o.status === "completed" || o.status === "cancelled" ? (
                  <span className="text-emerald-600 font-medium">+¥{o.price}</span>
                ) : (
                  <span className={`text-[10px] px-2 py-0.5 rounded-full ${o.status in statusColor ? statusColor[o.status] : "bg-gray-100 text-gray-600"}`}>
                    {statusLabel[o.status]}
                  </span>
                )}
              </div>
            </div>
            <div className="flex gap-2 text-xs mb-0.5 items-center">
              <div className="w-1.5 h-1.5 rounded-full bg-emerald-500 flex-shrink-0" />
              <span className="flex-1 text-gray-700 truncate">{o.from}</span>
            </div>
            <div className="flex gap-2 text-xs items-center">
              <div className="w-1.5 h-1.5 rounded-full bg-rose-500 flex-shrink-0" />
              <span className="flex-1 text-gray-700 truncate">{o.to}</span>
            </div>
            <div className="flex justify-between items-center mt-2.5 pt-2.5 border-t border-gray-50 text-[10px] text-gray-400">
              <span>{o.carType} · {o.distanceKm}km · {o.estMinutes}分钟</span>
              <ChevronRight className="w-3.5 h-3.5 text-gray-300" />
            </div>
          </button>
        ))}
        {list.length === 0 && (
          <div className="text-center py-12 text-gray-400 text-sm">
            <div className="text-3xl mb-2">📋</div>暂无订单记录
          </div>
        )}
      </div>
    </div>
  );
}

/* ============================================================
   行程详情（司机视角）
   ============================================================ */
function DriverTripDetailView({ order, onBack, onToast }: { order: Order; onBack: () => void; onToast: (m: string) => void }) {
  const timeline = [
    { time: new Date(order.createdAt).toLocaleTimeString("zh-CN", { hour: "2-digit", minute: "2-digit" }), label: "乘客下单", done: true },
    { time: "已接单", label: "接受订单", done: !!order.driverId },
    { time: "已到达", label: "到达上车点", done: ["arrived", "ongoing", "completed"].includes(order.status) },
    { time: "行程中", label: "开始行程", done: ["ongoing", "completed"].includes(order.status) },
    { time: "已完成", label: "完成行程", done: order.status === "completed" },
  ];

  return (
    <div className="min-h-full bg-gray-50">
      {/* 顶栏 */}
      <div className="bg-white px-4 py-3 flex items-center border-b sticky top-0 z-10">
        <button onClick={onBack}><ChevronLeft className="w-5 h-5" /></button>
        <div className="flex-1 text-center font-medium">行程详情</div>
        <button onClick={() => onToast("已复制订单号")} className="text-xs text-gray-400">
          <Copy className="w-4 h-4" />
        </button>
      </div>

      {/* 地图 */}
      <AmapView className="h-40" from={order.from} to={order.to} fromLat={order.fromLat} fromLng={order.fromLng} toLat={order.toLat} toLng={order.toLng} />

      {/* 状态 + 收入 */}
      <div className="mx-3 -mt-4 relative z-10 bg-gradient-to-r from-emerald-500 to-teal-500 rounded-2xl p-4 text-white shadow-lg mb-3">
        <div className="flex items-center justify-between mb-2">
          <span className={`text-[10px] px-2 py-0.5 rounded-full bg-white/20`}>{statusLabel[order.status]}</span>
          <span className="text-xs opacity-80">{new Date(order.createdAt).toLocaleString("zh-CN")}</span>
        </div>
        <div className="text-3xl font-light mb-1">+¥{order.price}</div>
        <div className="text-xs opacity-80">订单号：{order.orderNo || order.id}</div>
      </div>

      {/* 路线 */}
      <div className="mx-3 bg-white rounded-2xl p-4 shadow-sm mb-3">
        <div className="text-xs font-medium text-gray-500 mb-3">行程路线</div>
        <div className="space-y-3">
          <div className="flex gap-3 items-start">
            <div className="w-2 h-2 rounded-full bg-emerald-500 mt-1 flex-shrink-0" />
            <div><div className="text-sm text-gray-800">{order.from}</div><div className="text-[10px] text-gray-400">上车点</div></div>
          </div>
          <div className="flex gap-3 items-start">
            <div className="w-2 h-2 rounded-full bg-rose-500 mt-1 flex-shrink-0" />
            <div><div className="text-sm text-gray-800">{order.to}</div><div className="text-[10px] text-gray-400">下车点</div></div>
          </div>
        </div>
        <div className="flex gap-4 mt-3 pt-3 border-t border-gray-50 text-xs text-gray-500">
          <span>📏 {order.distanceKm}km</span>
          <span>⏱ {order.estMinutes}分钟</span>
          <span>🚗 {order.carType}</span>
        </div>
      </div>

      {/* 乘客信息 */}
      <div className="mx-3 bg-white rounded-2xl p-4 shadow-sm mb-3">
        <div className="text-xs font-medium text-gray-500 mb-3">乘客信息</div>
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-full bg-violet-100 flex items-center justify-center text-xl">👤</div>
          <div className="flex-1">
            <div className="text-sm font-medium">{order.passengerName}</div>
            <div className="text-[10px] text-gray-400">{order.passengerPhone}</div>
          </div>
          {order.rating && (
            <div className="text-right">
              <div className="flex items-center gap-0.5 justify-end">
                {[1, 2, 3, 4, 5].map(i => (
                  <Star key={i} className={`w-3 h-3 ${i <= order.rating! ? "text-amber-400 fill-current" : "text-gray-200"}`} />
                ))}
              </div>
              {order.ratingComment && <div className="text-[10px] text-gray-500 mt-0.5">"{order.ratingComment}"</div>}
              {order.ratingTags?.map((t: string) => (
                <span key={t} className="text-[9px] bg-amber-50 text-amber-600 px-1.5 py-0.5 rounded mr-1">{t}</span>
              ))}
            </div>
          )}
        </div>
        {order.note && <div className="mt-2 text-[10px] text-orange-500 bg-orange-50 rounded-lg px-2 py-1">备注：{order.note}</div>}
      </div>

      {/* 收入明细 */}
      <div className="mx-3 bg-white rounded-2xl p-4 shadow-sm mb-3">
        <div className="text-xs font-medium text-gray-500 mb-3">收入明细</div>
        {(order.baseFee || order.distanceFee || order.durationFee || order.waitFee) ? (
          <>
            <div className="flex justify-between py-2 border-b">
              <span className="text-sm text-gray-600">基础费</span>
              <span className="text-sm font-medium text-gray-800">¥{(order.baseFee ?? 0).toFixed(2)}</span>
            </div>
            <div className="flex justify-between py-2 border-b">
              <span className="text-sm text-gray-600">里程费</span>
              <span className="text-sm font-medium text-gray-800">¥{(order.distanceFee ?? 0).toFixed(2)}</span>
            </div>
            <div className="flex justify-between py-2 border-b">
              <span className="text-sm text-gray-600">时长费</span>
              <span className="text-sm font-medium text-gray-800">¥{(order.durationFee ?? 0).toFixed(2)}</span>
            </div>
            {order.waitFee != null && order.waitFee > 0 && (
              <div className="flex justify-between py-2 border-b">
                <span className="text-sm text-gray-600">等待费</span>
                <span className="text-sm font-medium text-gray-800">¥{(order.waitFee ?? 0).toFixed(2)}</span>
              </div>
            )}
            <div className="flex justify-between py-2">
              <span className="text-sm text-gray-600">合计</span>
              <span className="text-sm font-medium text-emerald-600">¥{order.price.toFixed(2)}</span>
            </div>
          </>
        ) : (
          <>
            <div className="flex justify-between py-2 border-b">
              <span className="text-sm text-gray-600">行程费用</span>
              <span className="text-sm font-medium text-gray-800">¥{order.price}</span>
            </div>
          </>
        )}
        {order.paymentMethod && (
          <div className="flex justify-between py-2 text-xs text-gray-400">
            <span>支付方式</span><span>{order.paymentMethod}</span>
          </div>
        )}
      </div>

      {/* 行程时间线 */}
      <div className="mx-3 bg-white rounded-2xl p-4 shadow-sm mb-3">
        <div className="text-xs font-medium text-gray-500 mb-3">行程轨迹</div>
        <div className="space-y-3">
          {timeline.map((t, i) => (
            <div key={i} className="flex items-center gap-3">
              <div className={`w-5 h-5 rounded-full flex items-center justify-center flex-shrink-0 ${t.done ? "bg-emerald-500" : "bg-gray-100"}`}>
                {t.done ? <CheckCircle2 className="w-3 h-3 text-white" /> : <div className="w-2 h-2 rounded-full bg-gray-300" />}
              </div>
              <div className="flex-1">
                <span className={`text-sm ${t.done ? "text-gray-800" : "text-gray-400"}`}>{t.label}</span>
              </div>
              <span className="text-[10px] text-gray-400">{t.time}</span>
            </div>
          ))}
        </div>
      </div>

      {/* 操作按钮 */}
      <div className="mx-3 mb-6 grid grid-cols-2 gap-3">
        <button onClick={() => onToast("开具电子发票")} className="bg-white rounded-2xl py-3 text-sm text-blue-500 border border-blue-100 shadow-sm flex items-center justify-center gap-2">
          <FileText className="w-4 h-4" />开具发票
        </button>
        <button onClick={() => onToast("已发起投诉")} className="bg-white rounded-2xl py-3 text-sm text-gray-500 border border-gray-100 shadow-sm flex items-center justify-center gap-2">
          <AlertTriangle className="w-4 h-4" />问题反馈
        </button>
      </div>
    </div>
  );
}

/* ============================================================
   司机个人中心
   ============================================================ */
function DriverMe({ driver, onNav, onWithdraw, onToast }: any) {
  return (
    <div>
      <div className="bg-gradient-to-b from-emerald-500 to-teal-600 px-4 pt-6 pb-8 relative overflow-hidden">
        <div className="absolute -right-4 -top-4 w-32 h-32 rounded-full bg-white/10" />
        <div className="absolute right-8 bottom-0 w-20 h-20 rounded-full bg-white/5" />
        <div className="flex items-center gap-3 mb-5 relative">
          <div className="w-16 h-16 rounded-full bg-white/30 border-2 border-white/50 flex items-center justify-center text-3xl">👨‍✈️</div>
          <div className="flex-1">
            <div className="text-white font-medium text-lg flex items-center gap-2">
              {driver.name}
              <span className="bg-amber-400 text-gray-800 text-[10px] px-1.5 py-0.5 rounded font-bold">金牌司机</span>
            </div>
            <div className="text-white/80 text-xs mt-1">{driver.plate} · {driver.car}</div>
          </div>
          <button onClick={() => onNav("settings")} className="w-9 h-9 rounded-full bg-white/20 flex items-center justify-center">
            <Settings className="w-4 h-4 text-white" />
          </button>
        </div>
        <div className="grid grid-cols-3 gap-2 relative">
          {[{ val: driver.rating, label: "综合评分" }, { val: driver.totalOrders, label: "总完单" }, { val: "98%", label: "好评率" }]
            .map(item => (
              <div key={item.label} className="bg-white/20 rounded-xl py-3 text-center text-white">
                <div className="text-lg font-medium">{item.val}</div>
                <div className="text-[10px] opacity-80 mt-0.5">{item.label}</div>
              </div>
            ))}
        </div>
      </div>

      {/* 收入卡 */}
      <div className="mx-3 -mt-4 relative z-10 bg-white rounded-2xl shadow-md p-4 flex items-center justify-between mb-3">
        <div>
          <div className="text-xs text-gray-500">今日可结收入</div>
          <div className="text-emerald-600 font-medium text-xl mt-0.5">¥{driver.todayEarnings.toFixed(2)}</div>
        </div>
        <div className="flex gap-2">
          <button onClick={() => onNav("income")} className="bg-gray-100 text-gray-600 text-xs px-3 py-2 rounded-full">查看明细</button>
          <button onClick={onWithdraw} className="bg-emerald-500 text-white text-xs px-3 py-2 rounded-full">立即提现</button>
        </div>
      </div>

      <div className="px-3 pb-6">
        <div className="bg-white rounded-2xl overflow-hidden">
          {[
            { icon: DollarSign, label: "收入明细", val: `¥${(driver.todayEarnings * 12).toFixed(0)}`, key: "income", color: "text-emerald-500" },
            { icon: TrendingUp, label: "服务分", val: `${Math.floor(driver.rating * 20)}/100`, key: "service", color: "text-blue-500" },
            { icon: Award, label: "司机等级", val: "金牌 Lv.3", key: null, color: "text-amber-500" },
            { icon: Car, label: "我的车辆", val: driver.car.split("·")[0].trim(), key: "car", color: "text-violet-500" },
            { icon: Shield, label: "安全中心", val: "", key: "sos", color: "text-teal-500" },
            { icon: Gift, label: "司机福利", val: "", key: null, color: "text-rose-500" },
            { icon: Settings, label: "设置", val: "", key: "settings", color: "text-gray-500" },
            { icon: Headphones, label: "联系客服", val: "", key: null, color: "text-indigo-500" },
          ].map(({ icon: Icon, label, val, key, color }) => (
            <button key={label} onClick={() => key ? onNav(key) : onToast(label)}
              className="w-full flex items-center gap-3 px-4 py-3.5 border-b border-gray-50 last:border-0 hover:bg-gray-50">
              <div className={`w-9 h-9 rounded-full bg-gray-50 flex items-center justify-center ${color}`}>
                <Icon className="w-4 h-4" />
              </div>
              <span className="flex-1 text-left text-sm text-gray-900">{label}</span>
              {val && <span className="text-xs text-gray-400">{val}</span>}
              <ChevronRight className="w-4 h-4 text-gray-300" />
            </button>
          ))}
        </div>
      </div>
    </div>
  );
}

/* ============================================================
   收入明细页
   ============================================================ */
function DriverIncomeView({ driver, onBack, onWithdraw, onToast }: any) {
  const [period, setPeriod] = useState<"today" | "week" | "month">("today");
  const periodData: any = {
    today: { income: driver.todayEarnings, trips: 8, hours: 6.2, base: driver.todayEarnings * 0.85, bonus: driver.todayEarnings * 0.1, subsidy: driver.todayEarnings * 0.05 },
    week: { income: driver.todayEarnings * 5.2, trips: 42, hours: 31, base: driver.todayEarnings * 5.2 * 0.85, bonus: driver.todayEarnings * 5.2 * 0.1, subsidy: driver.todayEarnings * 5.2 * 0.05 },
    month: { income: driver.todayEarnings * 22, trips: 183, hours: 132, base: driver.todayEarnings * 22 * 0.85, bonus: driver.todayEarnings * 22 * 0.1, subsidy: driver.todayEarnings * 22 * 0.05 },
  };
  const d = periodData[period];
  const bars = [42, 65, 58, 73, 89, 95, 78];

  return (
    <div className="min-h-full bg-gray-50">
      <div className="bg-white px-4 py-3 flex items-center border-b">
        <button onClick={onBack}><ChevronLeft className="w-5 h-5" /></button>
        <div className="flex-1 text-center font-medium">收入明细</div>
        <button onClick={onWithdraw} className="text-emerald-600 text-sm">提现</button>
      </div>

      <div className="bg-gradient-to-r from-emerald-500 to-teal-500 p-5 text-white text-center">
        <div className="text-xs opacity-90 mb-1">总收入 (元)</div>
        <div className="text-4xl font-light mb-3">¥{d.income.toFixed(2)}</div>
        <div className="flex justify-center gap-2 mb-4">
          {[["today", "今日"], ["week", "本周"], ["month", "本月"]].map(([k, l]) => (
            <button key={k} onClick={() => setPeriod(k as any)}
              className={`px-4 py-1.5 rounded-full text-xs ${period === k ? "bg-white text-emerald-600 font-medium" : "bg-white/20"}`}>{l}</button>
          ))}
        </div>
        <div className="grid grid-cols-3 gap-2 text-center text-[11px]">
          {[{ v: `${d.trips}单`, l: "完单数" }, { v: `${d.hours}h`, l: "出车时长" }, { v: `¥${(d.income / d.hours).toFixed(1)}`, l: "时均收入" }]
            .map(x => (
              <div key={x.l} className="bg-white/15 rounded-xl py-2">
                <div className="font-medium">{x.v}</div>
                <div className="opacity-80 mt-0.5">{x.l}</div>
              </div>
            ))}
        </div>
      </div>

      {/* 趋势图 */}
      <div className="mx-3 mt-3 bg-white rounded-2xl p-4 shadow-sm">
        <div className="text-sm font-medium mb-3">近7日收入趋势</div>
        <div className="flex items-end gap-1.5 h-24">
          {bars.map((h, i) => (
            <div key={i} className="flex-1 flex flex-col items-center gap-1">
              <div className="w-full rounded-t bg-gradient-to-t from-emerald-500 to-teal-400 min-h-[4px]"
                style={{ height: `${h * 0.9}px` }} />
              <div className="text-[9px] text-gray-400">{["一", "二", "三", "四", "五", "六", "日"][i]}</div>
            </div>
          ))}
        </div>
      </div>

      {/* 收入构成 */}
      <div className="mx-3 mt-3 bg-white rounded-2xl p-4 shadow-sm">
        <div className="text-sm font-medium mb-3">收入构成</div>
        {[
          { name: "行程基础收入", val: d.base, color: "bg-emerald-500", pct: 85 },
          { name: "奖励收入", val: d.bonus, color: "bg-amber-500", pct: 10 },
          { name: "活动补贴", val: d.subsidy, color: "bg-blue-500", pct: 5 },
        ].map(item => (
          <div key={item.name} className="py-2.5 border-b last:border-0">
            <div className="flex justify-between mb-1.5">
              <div className="flex items-center gap-2">
                <div className={`w-2 h-2 rounded-full ${item.color}`} />
                <span className="text-sm text-gray-700">{item.name}</span>
              </div>
              <span className="text-sm font-medium text-gray-900">¥{item.val.toFixed(2)}</span>
            </div>
            <div className="h-1.5 bg-gray-100 rounded-full overflow-hidden ml-4">
              <div className={`h-full ${item.color} rounded-full`} style={{ width: `${item.pct}%` }} />
            </div>
          </div>
        ))}
      </div>

      {/* 结算记录 */}
      <div className="mx-3 mt-3 mb-6 bg-white rounded-2xl p-4 shadow-sm">
        <div className="text-sm font-medium mb-3">最近结算记录</div>
        {[
          { date: "2026-04-22", amount: 289.5, status: "已结算" },
          { date: "2026-04-21", amount: 312.0, status: "已结算" },
          { date: "2026-04-20", amount: 245.8, status: "已结算" },
        ].map((r, i) => (
          <div key={i} className="flex justify-between items-center py-2.5 border-b last:border-0">
            <div className="text-sm text-gray-800">{r.date}</div>
            <div className="flex items-center gap-3">
              <span className="text-[10px] bg-emerald-100 text-emerald-700 px-2 py-0.5 rounded">{r.status}</span>
              <span className="text-emerald-600 font-medium">¥{r.amount.toFixed(2)}</span>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

/* ============================================================
   服务分页
   ============================================================ */
function DriverServiceView({ driver, onBack, onToast }: any) {
  const score = Math.floor(driver.rating * 20);
  const items = [
    { name: "行程评分", score: 98, desc: "基于乘客评价综合计算" },
    { name: "接单率", score: 94, desc: "分配订单中成功接单比例" },
    { name: "完单率", score: 99, desc: "接单后成功完成比例" },
    { name: "安全驾驶", score: 96, desc: "基于行程轨迹分析" },
  ];

  return (
    <div className="min-h-full bg-gray-50">
      <div className="bg-white px-4 py-3 flex items-center border-b">
        <button onClick={onBack}><ChevronLeft className="w-5 h-5" /></button>
        <div className="flex-1 text-center font-medium">服务分</div>
        <div className="w-5" />
      </div>

      <div className="bg-gradient-to-br from-blue-500 to-indigo-600 p-6 text-white text-center">
        <div className="text-xs opacity-80 mb-2">综合服务分</div>
        <div className="text-6xl font-bold mb-1">{score}</div>
        <div className="text-sm opacity-90">{score >= 90 ? "🏆 优秀" : score >= 80 ? "👍 良好" : "📊 待提升"}</div>
        <div className="mt-4 text-xs opacity-80">服务分影响您的派单优先级和奖励金额</div>
      </div>

      <div className="mx-3 mt-3 bg-white rounded-2xl p-4 shadow-sm">
        <div className="text-sm font-medium mb-3">分项明细</div>
        {items.map(item => (
          <div key={item.name} className="py-3 border-b last:border-0">
            <div className="flex justify-between items-center mb-1.5">
              <span className="text-sm text-gray-800">{item.name}</span>
              <span className="text-sm font-medium text-blue-600">{item.score}<span className="text-gray-400 text-xs">/100</span></span>
            </div>
            <div className="h-1.5 bg-gray-100 rounded-full overflow-hidden mb-1">
              <div className="h-full bg-gradient-to-r from-blue-400 to-indigo-500 rounded-full" style={{ width: `${item.score}%` }} />
            </div>
            <div className="text-[10px] text-gray-400">{item.desc}</div>
          </div>
        ))}
      </div>

      <div className="mx-3 mt-3 mb-6 bg-amber-50 border border-amber-200 rounded-2xl p-4">
        <div className="text-sm font-medium text-amber-700 mb-2 flex items-center gap-1.5">
          <Zap className="w-4 h-4" /> 提升建议
        </div>
        <div className="space-y-2 text-xs text-amber-600">
          {["提高接单率：保持良好在线率，及时响应派单", "主动服务：主动与乘客沟通，提升体验", "安全驾驶：避免急刹急加速，保持平稳行驶"]
            .map(t => (
              <div key={t} className="flex items-start gap-2">
                <CheckCircle2 className="w-3.5 h-3.5 mt-0.5 flex-shrink-0" />{t}
              </div>
            ))}
        </div>
      </div>
    </div>
  );
}

/* ============================================================
   车辆信息页
   ============================================================ */
function DriverCarView({ driver, onBack, onToast }: any) {
  return (
    <div className="min-h-full bg-gray-50">
      <div className="bg-white px-4 py-3 flex items-center border-b">
        <button onClick={onBack}><ChevronLeft className="w-5 h-5" /></button>
        <div className="flex-1 text-center font-medium">我的车辆</div>
        <button onClick={() => onToast("申请修改")} className="text-blue-500 text-sm">修改</button>
      </div>

      <div className="mx-3 mt-4 bg-gradient-to-br from-gray-800 to-gray-900 rounded-2xl p-5 text-white relative overflow-hidden">
        <div className="text-3xl mb-1">🚗</div>
        <div className="text-xl font-bold">{driver.plate}</div>
        <div className="text-sm opacity-80 mt-1">{driver.car}</div>
        <div className="absolute -right-3 -bottom-3 text-8xl opacity-10">🚗</div>
      </div>

      <div className="mx-3 mt-3 bg-white rounded-2xl overflow-hidden shadow-sm">
        {[
          { label: "车牌号", val: driver.plate },
          { label: "车辆品牌", val: driver.car.split("·")[0].trim() },
          { label: "车身颜色", val: driver.car.split("·")[1]?.trim() || "银色" },
          { label: "座位数", val: "5座" },
          { label: "排放标准", val: "国六" },
          { label: "营运证号", val: "运营J2026042" },
          { label: "保险到期", val: "2027-03-31" },
          { label: "年检到期", val: "2027-06-30" },
        ].map((item, i, arr) => (
          <div key={item.label} className={`flex items-center px-4 py-3.5 ${i < arr.length - 1 ? "border-b border-gray-50" : ""}`}>
            <span className="flex-1 text-sm text-gray-500">{item.label}</span>
            <span className="text-sm text-gray-900">{item.val}</span>
          </div>
        ))}
      </div>

      <div className="mx-3 mt-3 mb-6 grid grid-cols-2 gap-3">
        <button onClick={() => onToast("查看年检记录")} className="bg-white rounded-2xl py-3.5 text-sm text-blue-500 border border-blue-100 shadow-sm">年检记录 ›</button>
        <button onClick={() => onToast("保险续保")} className="bg-white rounded-2xl py-3.5 text-sm text-emerald-500 border border-emerald-100 shadow-sm">保险续保 ›</button>
      </div>
    </div>
  );
}

/* ============================================================
   安全中心
   ============================================================ */
function DriverSOSView({ onBack, onToast }: { onBack: () => void; onToast: (m: string) => void }) {
  const [recording, setRecording] = useState(false);
  return (
    <div className="min-h-full bg-gray-50">
      <div className="bg-white px-4 py-3 flex items-center border-b">
        <button onClick={onBack}><ChevronLeft className="w-5 h-5" /></button>
        <div className="flex-1 text-center font-medium">安全中心</div>
        <div className="w-5" />
      </div>

      {/* SOS紧急按钮 */}
      <div className="bg-gradient-to-br from-rose-500 to-red-600 mx-3 mt-4 rounded-2xl p-5 text-white text-center">
        <button onClick={() => onToast("已发出SOS求助信号！正在联系紧急联系人...")}
          className="w-20 h-20 rounded-full bg-white/20 border-4 border-white/50 flex items-center justify-center mx-auto mb-3 active:scale-95 transition-transform">
          <span className="text-white font-bold text-2xl">SOS</span>
        </button>
        <div className="text-sm font-medium mb-1">紧急求助</div>
        <div className="text-[11px] opacity-80">遇到危险时按下，将通知紧急联系人和平台</div>
      </div>

      {/* 快捷联系 */}
      <div className="mx-3 mt-3 bg-white rounded-2xl overflow-hidden shadow-sm">
        <div className="px-4 py-3 border-b">
          <div className="text-sm font-medium">紧急联系</div>
        </div>
        {[
          { icon: "🚔", label: "报警 110", sub: "公安机关紧急求助", action: () => onToast("正在拨打110...") },
          { icon: "🚑", label: "急救 120", sub: "医疗紧急救援", action: () => onToast("正在拨打120...") },
          { icon: "🎧", label: "平台紧急客服", sub: "7×24小时专属通道", action: () => onToast("正在接入客服...") },
        ].map(item => (
          <button key={item.label} onClick={item.action}
            className="w-full flex items-center gap-3 px-4 py-3.5 border-b last:border-0 hover:bg-gray-50">
            <div className="w-10 h-10 rounded-full bg-rose-50 flex items-center justify-center text-xl">{item.icon}</div>
            <div className="flex-1 text-left">
              <div className="text-sm font-medium text-gray-900">{item.label}</div>
              <div className="text-[10px] text-gray-400 mt-0.5">{item.sub}</div>
            </div>
            <PhoneCall className="w-4 h-4 text-gray-300" />
          </button>
        ))}
      </div>

      {/* 行车记录 */}
      <div className="mx-3 mt-3 bg-white rounded-2xl p-4 shadow-sm">
        <div className="flex items-center justify-between mb-3">
          <div className="text-sm font-medium">行驶记录仪</div>
          <button onClick={() => { setRecording(!recording); onToast(recording ? "已停止录制" : "录制已开启"); }}
            className={`flex items-center gap-1.5 text-xs px-3 py-1.5 rounded-full ${recording ? "bg-rose-100 text-rose-600" : "bg-gray-100 text-gray-600"}`}>
            <Volume2 className="w-3 h-3" />{recording ? "录制中" : "开启录制"}
          </button>
        </div>
        <div className="text-[11px] text-gray-500 bg-gray-50 rounded-xl px-3 py-2">
          行程录制功能保护您和乘客的权益，录音文件将加密存储72小时
        </div>
      </div>

      {/* 安全知识 */}
      <div className="mx-3 mt-3 mb-6 bg-blue-50 border border-blue-100 rounded-2xl p-4">
        <div className="text-sm font-medium text-blue-700 mb-2">安全驾驶提示</div>
        {["行程中请勿使用手机", "不得绕路或不走最优路线", "出现纠纷请第一时间联系平台", "遇乘客异常行为保持冷静，及时上报"].map(t => (
          <div key={t} className="flex items-start gap-2 text-xs text-blue-600 mb-1.5">
            <CheckCircle2 className="w-3.5 h-3.5 mt-0.5 flex-shrink-0 text-blue-500" />{t}
          </div>
        ))}
      </div>
    </div>
  );
}

/* ============================================================
   设置页
   ============================================================ */
function DriverSettingsView({ driver, onBack, onToggleOnline, onToast }: any) {
  const [sound, setSound] = useState(true);
  const [autoNav, setAutoNav] = useState(true);
  const [nightMode, setNightMode] = useState(false);

  return (
    <div className="min-h-full bg-gray-50">
      <div className="bg-white px-4 py-3 flex items-center border-b">
        <button onClick={onBack}><ChevronLeft className="w-5 h-5" /></button>
        <div className="flex-1 text-center font-medium">设置</div>
        <div className="w-5" />
      </div>

      {/* 在线状态 */}
      <div className="mx-3 mt-4 bg-white rounded-2xl overflow-hidden shadow-sm">
        <div className="px-4 py-3 border-b text-xs text-gray-400 font-medium">出车设置</div>
        <div className="flex items-center px-4 py-3.5 border-b">
          <div className="flex-1">
            <div className="text-sm text-gray-900">出车状态</div>
            <div className="text-[10px] text-gray-400 mt-0.5">{driver.online ? "当前出车中，可接收订单" : "已收车"}</div>
          </div>
          <button onClick={onToggleOnline}
            className={`relative w-12 h-6 rounded-full transition-colors ${driver.online ? "bg-emerald-500" : "bg-gray-200"}`}>
            <div className={`absolute top-0.5 w-5 h-5 rounded-full bg-white shadow transition-transform ${driver.online ? "translate-x-6" : "translate-x-0.5"}`} />
          </button>
        </div>
      </div>

      {/* 通知设置 */}
      <div className="mx-3 mt-3 bg-white rounded-2xl overflow-hidden shadow-sm">
        <div className="px-4 py-3 border-b text-xs text-gray-400 font-medium">通知与声音</div>
        {[
          { label: "新订单提示音", sub: "收到新订单时播放提示", val: sound, set: setSound },
          { label: "自动导航", sub: "接单后自动启动导航", val: autoNav, set: setAutoNav },
          { label: "夜间模式", sub: "23:00-06:00自动降低亮度", val: nightMode, set: setNightMode },
        ].map(item => (
          <div key={item.label} className="flex items-center px-4 py-3.5 border-b last:border-0">
            <div className="flex-1">
              <div className="text-sm text-gray-900">{item.label}</div>
              <div className="text-[10px] text-gray-400 mt-0.5">{item.sub}</div>
            </div>
            <button onClick={() => item.set(!item.val)}
              className={`relative w-12 h-6 rounded-full transition-colors ${item.val ? "bg-emerald-500" : "bg-gray-200"}`}>
              <div className={`absolute top-0.5 w-5 h-5 rounded-full bg-white shadow transition-transform ${item.val ? "translate-x-6" : "translate-x-0.5"}`} />
            </button>
          </div>
        ))}
      </div>

      {/* 账户 */}
      <div className="mx-3 mt-3 bg-white rounded-2xl overflow-hidden shadow-sm">
        <div className="px-4 py-3 border-b text-xs text-gray-400 font-medium">账户</div>
        {[
          { icon: Lock, label: "修改密码", action: () => onToast("修改密码") },
          { icon: Shield, label: "隐私设置", action: () => onToast("隐私设置") },
          { icon: FileText, label: "用户协议", action: () => onToast("用户协议") },
        ].map(({ icon: Icon, label, action }) => (
          <button key={label} onClick={action} className="w-full flex items-center px-4 py-3.5 border-b last:border-0 hover:bg-gray-50">
            <Icon className="w-4 h-4 text-gray-400 mr-3" />
            <span className="flex-1 text-left text-sm text-gray-900">{label}</span>
            <ChevronRight className="w-4 h-4 text-gray-300" />
          </button>
        ))}
      </div>

      <div className="mx-3 mt-3 mb-6">
        <button onClick={() => onToast("已退出登录")} className="w-full bg-white rounded-2xl py-3.5 text-sm text-rose-500 border border-rose-100 shadow-sm flex items-center justify-center gap-2">
          <LogOut className="w-4 h-4" />退出登录
        </button>
      </div>
    </div>
  );
}

/* ============================================================
   浮动新订单提醒
   ============================================================ */
function FloatingNewOrderAlert({ order, onAccept, onView }: { order: Order; onAccept: () => void; onView: () => void }) {
  return (
    <div className="absolute bottom-16 left-3 right-3 bg-white rounded-2xl shadow-2xl border border-orange-200 z-40 overflow-hidden">
      <div className="bg-gradient-to-r from-orange-500 to-amber-500 px-3 py-1.5 flex items-center justify-between">
        <div className="flex items-center gap-1.5 text-white text-xs font-medium">
          <Activity className="w-3.5 h-3.5" />新订单抢单中
        </div>
        <span className="text-white text-xs opacity-80">{order.carType}</span>
      </div>
      <div className="p-3">
        <div className="flex justify-between items-center mb-2">
          <div className="flex gap-2 text-xs flex-col flex-1 mr-2">
            <div className="flex gap-1.5 items-center">
              <div className="w-1.5 h-1.5 rounded-full bg-emerald-500 flex-shrink-0" />
              <span className="text-gray-700 truncate">{order.from}</span>
            </div>
            <div className="flex gap-1.5 items-center">
              <div className="w-1.5 h-1.5 rounded-full bg-rose-500 flex-shrink-0" />
              <span className="text-gray-700 truncate">{order.to}</span>
            </div>
          </div>
          <div className="text-orange-600 font-bold text-lg">¥{order.price}</div>
        </div>
        <SlideToConfirm label="抢单" emoji="🚗" gradient="from-orange-500 to-amber-500"
          onConfirm={onAccept} resetKey={order.id + "-float"} />
        <button onClick={onView} className="w-full text-center text-xs text-gray-400 mt-2">查看详情 ›</button>
      </div>
    </div>
  );
}

/* ============================================================
   派单确认弹窗
   ============================================================ */
function DispatchConfirmModal({ order, onAccept, onReject }: { order: Order; onAccept: () => void; onReject: () => void }) {
  return (
    <div className="absolute inset-0 bg-black/50 z-50 flex items-center justify-center p-6">
      <div className="bg-white rounded-2xl p-5 w-full shadow-2xl">
        <div className="text-center mb-4">
          <div className="text-4xl mb-2">📢</div>
          <div className="font-medium text-gray-900">平台为您派单</div>
        </div>
        <div className="bg-orange-50 rounded-xl p-3 mb-4 space-y-2">
          <div className="flex items-start gap-2 text-sm">
            <div className="w-2 h-2 rounded-full bg-emerald-500 mt-1.5 flex-shrink-0" />
            <span className="text-gray-700 flex-1">{order.from}</span>
          </div>
          <div className="flex items-start gap-2 text-sm">
            <div className="w-2 h-2 rounded-full bg-rose-500 mt-1.5 flex-shrink-0" />
            <span className="text-gray-700 flex-1">{order.to}</span>
          </div>
          <div className="flex justify-between text-xs text-gray-500 pt-1">
            <span>{order.distanceKm}km · 约{order.estMinutes}分钟</span>
            <span className="text-orange-600 font-medium">¥{order.price}</span>
          </div>
        </div>
        <div className="flex gap-3">
          <button onClick={onReject} className="flex-1 py-3 rounded-full border border-gray-300 text-gray-700 font-medium">
            拒单
          </button>
          <button onClick={onAccept} className="flex-1 py-3 rounded-full bg-gradient-to-r from-orange-500 to-amber-500 text-white font-medium">
            接受
          </button>
        </div>
      </div>
    </div>
  );
}

/* ============================================================
   乘客取消弹窗
   ============================================================ */
function CancellationModal({ name, onClose }: { name: string; onClose: () => void }) {
  return (
    <div className="absolute inset-0 bg-black/50 z-50 flex items-center justify-center p-6">
      <div className="bg-white rounded-2xl p-5 w-full shadow-2xl">
        <div className="text-center mb-4">
          <div className="text-4xl mb-2">😔</div>
          <div className="font-medium text-gray-900">乘客已取消订单</div>
          <div className="text-sm text-gray-500 mt-1">{name} 已取消了本次行程</div>
        </div>
        <div className="bg-amber-50 rounded-xl p-3 text-xs text-amber-700 mb-4">
          取消费将在24小时内原路返还，若乘客有违规取消行为，平台将给予相应补偿。
        </div>
        <button onClick={onClose} className="w-full bg-gradient-to-r from-emerald-500 to-teal-500 text-white py-3 rounded-full font-medium">
          知道了，继续接单
        </button>
      </div>
    </div>
  );
}

/* ============================================================
   SOS 弹窗（行程中）
   ============================================================ */
function DriverSOSModal({ onClose, onToast }: { onClose: () => void; onToast: (m: string) => void }) {
  return (
    <div className="absolute inset-0 bg-black/60 z-50 flex items-end">
      <div className="bg-white rounded-t-3xl w-full p-5">
        <div className="flex items-center justify-between mb-4">
          <div className="font-medium text-gray-900 text-lg">紧急求助</div>
          <button onClick={onClose}><X className="w-5 h-5 text-gray-500" /></button>
        </div>
        <div className="grid grid-cols-3 gap-3 mb-4">
          {[
            { icon: "🚔", label: "报警 110", color: "bg-blue-50 border-blue-200", action: () => { onToast("正在拨打110..."); onClose(); } },
            { icon: "🚑", label: "急救 120", color: "bg-rose-50 border-rose-200", action: () => { onToast("正在拨打120..."); onClose(); } },
            { icon: "🎧", label: "紧急客服", color: "bg-violet-50 border-violet-200", action: () => { onToast("接入紧急客服..."); onClose(); } },
          ].map(item => (
            <button key={item.label} onClick={item.action}
              className={`${item.color} border rounded-2xl py-4 flex flex-col items-center gap-2`}>
              <span className="text-2xl">{item.icon}</span>
              <span className="text-xs text-gray-700">{item.label}</span>
            </button>
          ))}
        </div>
        <button onClick={() => { onToast("已向平台发送SOS！"); onClose(); }}
          className="w-full bg-rose-500 text-white py-3.5 rounded-full font-medium text-base">
          🆘 向平台发送SOS求助
        </button>
        <div className="text-[10px] text-center text-gray-400 mt-2">平台将立即响应并联系紧急联系人</div>
      </div>
    </div>
  );
}

/* ============================================================
   提现弹窗
   ============================================================ */
function WithdrawalModal({ balance, onClose, onToast }: { balance: number; onClose: () => void; onToast: (m: string) => void }) {
  const [amount, setAmount] = useState(String(balance.toFixed(2)));
  const [card] = useState("工商银行 **** 8821");

  function handleConfirm() {
    const v = parseFloat(amount);
    if (!v || v <= 0) { onToast("请输入有效金额"); return; }
    if (v > balance) { onToast("提现金额不能超过可用余额"); return; }
    onToast(`¥${v.toFixed(2)} 提现申请已提交，预计2小时到账`);
    onClose();
  }

  return (
    <div className="absolute inset-0 bg-black/50 z-50 flex items-end">
      <div className="bg-white rounded-t-3xl w-full p-5">
        <div className="flex items-center justify-between mb-4">
          <div className="font-medium text-gray-900">提现</div>
          <button onClick={onClose}><X className="w-5 h-5 text-gray-500" /></button>
        </div>

        <div className="bg-emerald-50 rounded-2xl p-3 mb-4 text-center">
          <div className="text-xs text-gray-500">可提现余额</div>
          <div className="text-3xl font-light text-emerald-600 mt-1">¥{balance.toFixed(2)}</div>
        </div>

        <div className="mb-3">
          <div className="text-xs text-gray-500 mb-1.5">提现金额</div>
          <div className="flex items-center border border-gray-200 rounded-xl px-3 py-2.5 focus-within:border-emerald-500">
            <span className="text-gray-400 mr-1.5">¥</span>
            <input type="number" value={amount} onChange={e => setAmount(e.target.value)}
              className="flex-1 outline-none text-lg font-medium text-gray-900" />
            <button onClick={() => setAmount(String(balance.toFixed(2)))} className="text-xs text-emerald-500">全部</button>
          </div>
        </div>

        <div className="flex items-center gap-2 bg-gray-50 rounded-xl px-3 py-2.5 mb-4">
          <CreditCard className="w-4 h-4 text-gray-400" />
          <span className="flex-1 text-sm text-gray-700">{card}</span>
          <ChevronDown className="w-4 h-4 text-gray-300" />
        </div>

        <div className="text-[10px] text-gray-400 mb-4">工作日2小时内到账，节假日顺延 · 单笔最低提现¥10</div>

        <button onClick={handleConfirm}
          className="w-full bg-gradient-to-r from-emerald-500 to-teal-500 text-white py-3.5 rounded-full font-medium">
          确认提现 ¥{parseFloat(amount) > 0 ? parseFloat(amount).toFixed(2) : "0.00"}
        </button>
      </div>
    </div>
  );
}

/* ============================================================
   通知面板
   ============================================================ */
function NotificationsPanel({ orders, onClose }: { orders: Order[]; onClose: () => void }) {
  const recents = orders.slice(0, 5);
  const systemNotifs = [
    { icon: "🎁", title: "本周完单奖励", body: "您本周完成18单，获得额外奖励¥50！", time: "10分钟前" },
    { icon: "⚡", title: "早高峰调度", body: "明日07:00-09:00早高峰补贴提高至1.5倍", time: "1小时前" },
    { icon: "📋", title: "服务分更新", body: "您的服务分已更新为98分，继续保持！", time: "3小时前" },
  ];

  return (
    <div className="absolute inset-0 bg-black/50 z-50 flex">
      <div className="flex-1" onClick={onClose} />
      <div className="w-[85%] bg-white h-full flex flex-col">
        <div className="px-4 py-3 border-b flex items-center justify-between">
          <div className="font-medium">消息通知</div>
          <button onClick={onClose}><X className="w-5 h-5 text-gray-500" /></button>
        </div>
        <div className="flex-1 overflow-y-auto">
          <div className="px-4 py-2 text-xs text-gray-400 font-medium mt-2">系统消息</div>
          {systemNotifs.map((n, i) => (
            <div key={i} className="flex gap-3 px-4 py-3 border-b hover:bg-gray-50">
              <div className="w-10 h-10 rounded-full bg-gray-100 flex items-center justify-center text-xl flex-shrink-0">{n.icon}</div>
              <div className="flex-1 min-w-0">
                <div className="text-sm font-medium text-gray-900">{n.title}</div>
                <div className="text-xs text-gray-500 mt-0.5">{n.body}</div>
                <div className="text-[10px] text-gray-400 mt-1">{n.time}</div>
              </div>
            </div>
          ))}
          <div className="px-4 py-2 text-xs text-gray-400 font-medium mt-2">近期完单</div>
          {recents.length === 0 && <div className="text-center text-gray-400 text-sm py-6">暂无记录</div>}
          {recents.map(o => (
            <div key={o.id} className="flex gap-3 px-4 py-3 border-b">
              <div className="w-10 h-10 rounded-full bg-emerald-100 flex items-center justify-center text-xl flex-shrink-0">✅</div>
              <div className="flex-1 min-w-0">
                <div className="text-sm text-gray-800 truncate">{o.from} → {o.to}</div>
                <div className="text-xs text-emerald-600 mt-0.5">+¥{o.price}</div>
                <div className="text-[10px] text-gray-400 mt-0.5">{new Date(o.createdAt).toLocaleString("zh-CN")}</div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
