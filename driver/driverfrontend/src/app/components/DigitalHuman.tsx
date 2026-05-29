import { useState, useRef, useEffect, useCallback } from "react";
import { Mic, X, Send, ChevronUp, Volume2, VolumeX, Ear } from "lucide-react";
import { useDigitalHuman } from "../utils/useDigitalHuman";
import type { SpeechState } from "../utils/speechEngine";
import { useStore } from "../store";
import ModelViewer from "./Model3D";

const DH_IMG = "/dh/img.png";
const QUICK_CMDS = ["今天接了几单？", "现在什么状态？", "你好"];

/* ============================================================
   悬浮气泡
   ============================================================ */
export function DigitalHumanBubble({ onClick, voiceMode, voiceState }: { onClick: () => void; voiceMode?: boolean; voiceState?: SpeechState }) {
  const isListen = voiceState === "listening";
  const isSpeak = voiceState === "speaking";
  const fast = isListen ? "0.8s" : isSpeak ? "1.5s" : "2.5s";
  const bright = isListen ? "0.5" : isSpeak ? "0.35" : "0.2";

  const glow = (w: number, blur: string, bg: string, delay: number) => (
    <div style={{
      position: "absolute", top: "50%", left: "50%",
      transform: "translate(-50%, -50%)",
      width: w, height: w, borderRadius: "50%",
      background: bg,
      animation: `rippleGlow ${fast} ease-out ${delay}s infinite`,
      filter: `blur(${blur})`, pointerEvents: "none",
    }} />
  );

  return (
    <button
      onClick={onClick}
      className={`absolute bottom-24 right-3 z-40 w-14 h-14 rounded-full shadow-lg cursor-pointer
        active:scale-95 transition-transform relative ${voiceMode ? "animate-pulse" : ""}`}
      style={{
        background: "linear-gradient(135deg, #f5a0b8, #f472b6)",
        boxShadow: voiceMode
          ? "0 0 24px rgba(244, 114, 182, 0.6), 0 0 60px rgba(191, 219, 254, 0.4)"
          : "0 0 20px rgba(244, 114, 182, 0.4), 0 0 40px rgba(191, 219, 254, 0.2)",
        animation: "dhFloat 3s ease-in-out infinite",
      }}
    >
      {/* 光晕层（相对按钮定位，跟随拖动） */}
      {voiceMode && glow(isListen ? 400 : 280, "12px",
        `radial-gradient(circle, rgba(191,219,254,${+bright * 0.4}) 0%, transparent 60%)`, +fast.replace("s","") * 0.6)}
      {voiceMode && glow(isListen ? 240 : 160, isListen ? "16px" : "10px",
        `radial-gradient(circle, rgba(244,114,182,${+bright * 0.5}) 0%, transparent 50%)`, +fast.replace("s","") * 0.3)}
      {voiceMode && glow(isListen ? 320 : 220, isListen ? "12px" : "6px",
        `radial-gradient(circle, rgba(244,114,182,${bright}) 0%, rgba(191,219,254,${+bright * 0.4}) 40%, transparent 70%)`, 0)}
      <div className="w-full h-full rounded-full overflow-hidden p-0.5 relative z-10">
        <img
          src={DH_IMG}
          alt="花小猪助手"
          className="w-full h-full object-cover rounded-full"
          style={{ objectPosition: "30% 20%" }}
        />
      </div>
      {/* 霓虹光环 */}
      <div className="absolute inset-0 rounded-full border-2 border-transparent pointer-events-none z-20"
        style={{
          background: "linear-gradient(135deg, rgba(191,219,254,0.6), rgba(244,114,182,0.3)) border-box",
          WebkitMask: "linear-gradient(#fff 0 0) padding-box, linear-gradient(#fff 0 0)",
          WebkitMaskComposite: "xor",
          maskComposite: "exclude",
        }}
      />
      <style>{`
        @keyframes dhFloat {
          0%, 100% { transform: translateY(0); }
          50% { transform: translateY(-6px); }
        }
      `}</style>
    </button>
  );
}

/* ============================================================
   全屏对话面板
   ============================================================ */
export function DigitalHumanPanel({ onClose }: { onClose: () => void }) {
  const store = useStore();
  const {
    state, messages, showWelcome, speechSupported,
    isFirstTime, startVoice, stopVoice, interrupt,
    handleUserInput, voiceMode, toggleVoiceMode,
    voicePer, setVoicePer, cleanup, dismissFirstTime,
  } = useDigitalHuman();

  const [textMode, setTextMode] = useState(false);
  const [inputText, setInputText] = useState("");
  const [showQuick, setShowQuick] = useState(false);
  const [panelReady, setPanelReady] = useState(false);
  const [modelFailed, setModelFailed] = useState(false);
  const [showVoicePicker, setShowVoicePicker] = useState(false);
  const [navigating, setNavigating] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const voicePickerRef = useRef<HTMLDivElement>(null);
  const pigClickTimer = useRef<ReturnType<typeof setTimeout> | null>(null);

  const handlePigClick = useCallback(() => {
    if (pigClickTimer.current) {
      clearTimeout(pigClickTimer.current);
      pigClickTimer.current = null;
      setShowVoicePicker(v => !v);
    } else {
      pigClickTimer.current = setTimeout(() => {
        pigClickTimer.current = null;
        toggleVoiceMode();
      }, 250);
    }
  }, [toggleVoiceMode]);

  const VOICES = [
    { per: "3", name: "度逍遥（男声）" },
    { per: "4", name: "度丫丫（女声）" },
    { per: "5003", name: "度逍遥（情感男声）" },
    { per: "5118", name: "度小鹿（情感女声）" },
  ];
  const inputRef = useRef<HTMLInputElement>(null);

  // 打开动画
  useEffect(() => {
    requestAnimationFrame(() => setPanelReady(true));
  }, []);

  // 自动滚动到底部
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  // 点击音色面板外部关闭
  useEffect(() => {
    if (!showVoicePicker) return;
    const handler = (e: MouseEvent) => {
      if (voicePickerRef.current && !voicePickerRef.current.contains(e.target as Node)) {
        setShowVoicePicker(false);
      }
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, [showVoicePicker]);

  // 首次使用延迟欢迎
  useEffect(() => {
    if (isFirstTime && panelReady) {
      const t = setTimeout(() => dismissFirstTime(), 600);
      return () => clearTimeout(t);
    }
  }, [isFirstTime, panelReady, dismissFirstTime]);

  // 关闭时清理引擎
  const handleClose = useCallback(() => {
    if (pigClickTimer.current) clearTimeout(pigClickTimer.current);
    cleanup();
    onClose();
  }, [cleanup, onClose]);

  // 发送文字
  const sendText = useCallback(() => {
    const text = inputText.trim();
    if (!text) return;
    setInputText("");
    handleUserInput(text);
  }, [inputText, handleUserInput]);

  // 点击快捷指令
  const handleQuick = useCallback((cmd: string) => {
    setShowQuick(false);
    handleUserInput(cmd);
  }, [handleUserInput]);

  // 处理文字输入框回车
  const handleKeyDown = useCallback((e: React.KeyboardEvent) => {
    if (e.key === "Enter") sendText();
  }, [sendText]);

  // 点击话筒
  const handleMicClick = useCallback(() => {
    if (state === "listening") {
      stopVoice();
    } else if (state === "speaking") {
      interrupt();
    } else if (state === "idle") {
      startVoice();
    }
  }, [state, startVoice, stopVoice, interrupt]);

  // 状态提示文字
  const getStatusHint = (): string => {
    switch (state) {
      case "listening": return "正在聆听...";
      case "thinking": return "让我想想...";
      case "speaking": return "";
      default: return "";
    }
  };

  // 话筒按钮样式
  const micBtnClass = () => {
    if (state === "listening") return "bg-red-500 scale-110 shadow-lg shadow-red-400/50";
    if (state === "speaking") return "bg-pink-400";
    if (!speechSupported) return "bg-gray-300 cursor-not-allowed";
    return "bg-pink-500 hover:bg-pink-400";
  };

  return (
    <div className="absolute inset-0 z-[70] flex flex-col overflow-hidden"
      style={{
        background: "linear-gradient(180deg, #fce4ec 0%, #fff5f7 40%, #ffffff 100%)",
        animation: panelReady ? "dhPanelIn 0.3s ease-out" : "none",
      }}
    >
      {/* 头部标题栏 */}
      <div className="flex items-center justify-between px-4 py-3 shrink-0 bg-gradient-to-r from-pink-500 to-pink-400">
        <button onClick={handleClose} className="text-white/90 hover:text-white p-1">
          <X className="w-5 h-5" />
        </button>
        <div className="text-white font-medium text-sm tracking-wide">花小猪助手</div>
        <button
          onClick={() => {
            if (navigating) return;
            setNavigating(true);
            store.setVoiceMode(true);
            setTimeout(() => handleClose(), 200);
          }}
          className={`flex items-center gap-1 text-xs rounded-full px-3 py-1 transition-all duration-200 select-none
            active:scale-95 active:brightness-90
            ${navigating
              ? "text-white/90 bg-white/15 animate-pulse shadow-[0_0_10px_rgba(255,255,255,0.3)]"
              : store.voiceMode
                ? "text-pink-600 bg-white font-medium"
                : "text-white/90 hover:text-white bg-white/15 hover:bg-white/25"
            }`}
          title="开启后回到接单页面可直接说话">
          {store.voiceMode && !navigating ? (
            <svg className="w-3.5 h-3.5" viewBox="0 0 16 16" fill="currentColor"><path d="M13.78 4.22a.75.75 0 0 1 0 1.06l-7 7a.75.75 0 0 1-1.06 0l-3.5-3.5a.75.75 0 1 1 1.06-1.06L6.25 10.69l6.47-6.47a.75.75 0 0 1 1.06 0z"/></svg>
          ) : (
            <Ear className="w-3.5 h-3.5" />
          )}
          {navigating ? "跳转中..." : store.voiceMode ? "语音已开启" : "语音助手"}
        </button>
      </div>

      {/* 上半区：3D 数字人展示 */}
      <div
        className="relative shrink-0 overflow-visible"
        style={{
          height: "40vh",
          minHeight: 200,
        }}
      >
        {/* 背景几何线装饰 */}
        <svg className="absolute inset-0 w-full h-full opacity-20 pointer-events-none" viewBox="0 0 400 300">
          <line x1="0" y1="80" x2="400" y2="80" stroke="#bae6fd" strokeWidth="1" strokeDasharray="8 4" />
          <line x1="0" y1="160" x2="400" y2="160" stroke="#fbcfe8" strokeWidth="1" strokeDasharray="4 8" />
          <line x1="0" y1="240" x2="400" y2="240" stroke="#bae6fd" strokeWidth="1" strokeDasharray="12 6" />
          <circle cx="350" cy="40" r="30" fill="none" stroke="#fbcfe8" strokeWidth="0.5" />
          <circle cx="50" cy="120" r="20" fill="none" stroke="#bae6fd" strokeWidth="0.5" />
          <rect x="10" y="10" width="80" height="60" rx="10" fill="none" stroke="#e0e0e0" strokeWidth="0.5" />
        </svg>

        {/* 粉色地面反光 */}
        <div className="absolute bottom-0 left-1/2 -translate-x-1/2 w-3/4 h-8 bg-pink-200/30 blur-xl rounded-full pointer-events-none" />

        {/* 3D 模型 / 2D 降级 */}
        {modelFailed ? (
          <div className="absolute inset-0 flex items-center justify-center">
            <img src={DH_IMG} alt="花小猪助手"
              className="w-[55%] max-w-[280px] h-auto object-contain drop-shadow-lg cursor-pointer"
              style={{
                animation: "dhIdleFloat 4s ease-in-out infinite",
              }}
              onClick={handlePigClick}
            />
          </div>
        ) : (
          <div className="absolute inset-0" onClick={handlePigClick}>
            <ModelViewer state={state} onError={() => setModelFailed(true)} />
          </div>
        )}

        {/* 语音模式开关指示 */}
        <button onClick={toggleVoiceMode}
          className={`absolute top-3 right-4 z-20 w-8 h-8 rounded-full flex items-center justify-center
            transition-all shadow-sm border ${
            voiceMode
              ? "bg-pink-500 text-white border-pink-400 shadow-pink-300/30"
              : "bg-white/80 text-gray-500 border-gray-200 hover:bg-white"
          }`}
          title={voiceMode ? "已开启语音回复，点击关闭" : "已关闭语音回复，点击开启"}>
          {voiceMode ? <Volume2 className="w-4 h-4" /> : <VolumeX className="w-4 h-4" />}
        </button>
        {/* 音色选择面板 */}
        {showVoicePicker && (
          <div ref={voicePickerRef}
            className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 z-30
              bg-white/95 backdrop-blur rounded-xl shadow-xl border border-gray-100 py-2 min-w-[170px]"
            onClick={e => e.stopPropagation()}>
            <div className="text-xs text-gray-400 px-4 pb-1.5 border-b border-gray-50">选择音色</div>
            {VOICES.map(v => (
              <button key={v.per} onClick={() => { setVoicePer(v.per); setShowVoicePicker(false); }}
                className={`w-full text-left px-4 py-2 text-sm transition-colors ${
                  voicePer === v.per ? "text-pink-600 bg-pink-50 font-medium" : "text-gray-700 hover:bg-gray-50"
                }`}>
                {voicePer === v.per ? "✓ " : "  "}{v.name}
              </button>
            ))}
          </div>
        )}
      </div>

      {/* 阴影分割 */}
      <div className="relative shrink-0 h-[10px] -mt-px pointer-events-none"
        style={{
          background: "linear-gradient(180deg, rgba(244,114,182,0.12) 0%, transparent 100%)",
        }}
      />

      {/* 下半区：对话记录 */}
      <div className="flex-1 flex flex-col min-h-0 bg-white/70 backdrop-blur-md rounded-t-lg shadow-[0_-2px_12px_rgba(244,114,182,0.06)]">
        {/* 状态提示 */}
        {state !== "idle" && (
          <div className="text-center text-xs text-pink-400 py-1.5 shrink-0">
            {getStatusHint()}
          </div>
        )}

        {/* 对话列表 */}
        <div className="flex-1 overflow-y-auto px-3 py-2 space-y-3">
          {showWelcome && messages.length === 0 && (
            <div className="flex flex-col items-center justify-center h-full text-gray-400 text-xs space-y-2">
              <img src={DH_IMG} alt="" className="w-12 h-12 rounded-full object-cover opacity-50"
                style={{ objectPosition: "30% 20%" }} />
              <div>点击下面的话筒开始对话</div>
            </div>
          )}

          {messages.map((msg, i) => (
            <div key={i} className={`flex gap-2 ${msg.role === "user" ? "flex-row-reverse" : ""}`}>
              {msg.role === "assistant" && (
                <img src={DH_IMG} alt="" className="w-8 h-8 rounded-full object-cover shrink-0 mt-0.5"
                  style={{ objectPosition: "30% 20%" }} />
              )}
              <div className={`max-w-[75%] px-3 py-2 rounded-2xl text-sm leading-relaxed ${
                msg.role === "user"
                  ? "bg-gray-100 text-gray-800 rounded-tr-sm"
                  : "bg-pink-50 text-gray-800 rounded-tl-sm"
              }`}>
                {msg.content}
              </div>
            </div>
          ))}
          <div ref={messagesEndRef} />
        </div>

        {/* 快捷指令 */}
        {showQuick && (
          <div className="shrink-0 px-3 py-2 flex gap-2 overflow-x-auto border-t border-gray-100">
            {QUICK_CMDS.map(cmd => (
              <button key={cmd} onClick={() => handleQuick(cmd)}
                className="shrink-0 text-xs px-3 py-1.5 rounded-full bg-pink-50 text-pink-600 border border-pink-200
                  hover:bg-pink-100 active:scale-95 transition-all">
                {cmd}
              </button>
            ))}
          </div>
        )}

        {/* 底部输入栏 */}
        <div className="shrink-0 border-t border-gray-100 bg-white px-3 py-2 flex items-center gap-2">
          {textMode ? (
            <>
              <input
                ref={inputRef}
                type="text"
                value={inputText}
                onChange={e => setInputText(e.target.value)}
                onKeyDown={handleKeyDown}
                placeholder="输入消息..."
                className="flex-1 outline-none text-sm bg-gray-50 rounded-full px-3 py-2"
                autoFocus
              />
              <button onClick={sendText}
                className="w-9 h-9 rounded-full bg-pink-500 text-white flex items-center justify-center
                  hover:bg-pink-400 active:scale-95 transition-all shrink-0">
                <Send className="w-4 h-4" />
              </button>
              <button onClick={() => setTextMode(false)}
                className="w-9 h-9 rounded-full bg-gray-100 text-gray-500 flex items-center justify-center shrink-0">
                <Mic className="w-4 h-4" />
              </button>
            </>
          ) : (
            <>
              <button onClick={() => setShowQuick(v => !v)}
                className={`px-3 py-1.5 rounded-full text-xs border transition-all shrink-0 ${
                  showQuick ? "bg-pink-50 border-pink-200 text-pink-600" : "bg-gray-50 border-gray-200 text-gray-500"
                }`}>
                <ChevronUp className="w-3.5 h-3.5 inline mr-0.5" />
                快捷
              </button>

              <button
                onClick={handleMicClick}
                disabled={!speechSupported}
                className={`w-12 h-12 rounded-full flex items-center justify-center transition-all duration-300 shrink-0 ${micBtnClass()}`}>
                <Mic className="w-5 h-5 text-white" />
              </button>

              <button onClick={() => { setTextMode(true); setShowQuick(false); }}
                className={`px-3 py-1.5 rounded-full text-xs border transition-all shrink-0 ${
                  state !== "idle" ? "bg-gray-100 border-gray-200 text-gray-300 cursor-not-allowed" : "bg-gray-50 border-gray-200 text-gray-500"
                }`}
                disabled={state !== "idle"}>
                文字
              </button>
            </>
          )}
        </div>
      </div>

      <style>{`
        @keyframes dhPanelIn {
          from { opacity: 0; transform: scale(0.95); }
          to { opacity: 1; transform: scale(1); }
        }
        @keyframes dhIdleFloat {
          0%, 100% { transform: translateY(0); }
          50% { transform: translateY(-8px); }
        }
        @keyframes dhBreathSlow {
          from { opacity: 0.4; }
          to { opacity: 0.9; }
        }
        @keyframes dhBreathFast {
          from { opacity: 0.3; }
          to { opacity: 1; }
        }
      `}</style>
    </div>
  );
}
