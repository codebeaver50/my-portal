"use client";

import { cloneElement, useCallback, useEffect, useId, useRef, useState } from "react";
import type { ReactElement } from "react";
import { createPortal } from "react-dom";

type TooltipProps = {
  content: string;
  // ツールチップを付ける要素。aria-describedby で content と関連付ける
  children: ReactElement<{ "aria-describedby"?: string }>;
};

type Position = { top: number; left: number; placement: "top" | "bottom" };

const GAP = 8;
const VIEWPORT_MARGIN = 8;

// children にホバー・フォーカス・タップでツールチップを出す。
// 表やモーダル等のスクロール領域で切れないよう、body へのポータルに fixed で配置する。
// ラッパーは display: contents にして、children のレイアウト（w-full 等）を変えない。
export function Tooltip({ content, children }: TooltipProps) {
  const id = useId();
  const wrapperRef = useRef<HTMLSpanElement>(null);
  const tooltipRef = useRef<HTMLSpanElement>(null);
  const [position, setPosition] = useState<Position | null>(null);

  const show = useCallback(() => {
    const trigger = wrapperRef.current?.firstElementChild;
    if (!trigger) return;
    const rect = trigger.getBoundingClientRect();
    // 上に出す余裕がなければ下に出す（実際の高さは表示後に補正する）
    const placement = rect.top > 48 ? "top" : "bottom";
    setPosition({
      top: placement === "top" ? rect.top - GAP : rect.bottom + GAP,
      left: rect.left + rect.width / 2,
      placement,
    });
  }, []);
  const hide = useCallback(() => setPosition(null), []);

  // 画面端からはみ出さないよう、表示後に横位置を補正する
  useEffect(() => {
    const tooltip = tooltipRef.current;
    if (!position || !tooltip) return;
    const rect = tooltip.getBoundingClientRect();
    let shift = 0;
    if (rect.left < VIEWPORT_MARGIN) shift = VIEWPORT_MARGIN - rect.left;
    if (rect.right > window.innerWidth - VIEWPORT_MARGIN) shift = window.innerWidth - VIEWPORT_MARGIN - rect.right;
    tooltip.style.marginLeft = `${shift}px`;
  }, [position]);

  // 表示中は、外側のタップ・Escキーで閉じる。スクロール時は閉じずに位置を合わせ直す
  // （Tabで画面外のボタンにフォーカスすると自動でスクロールされるため、閉じると読めなくなる）
  useEffect(() => {
    if (!position) return;
    function handlePointerDown(event: PointerEvent) {
      if (!wrapperRef.current?.contains(event.target as Node)) hide();
    }
    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === "Escape") hide();
    }
    document.addEventListener("pointerdown", handlePointerDown);
    document.addEventListener("keydown", handleKeyDown);
    window.addEventListener("scroll", show, true);
    window.addEventListener("resize", show);
    return () => {
      document.removeEventListener("pointerdown", handlePointerDown);
      document.removeEventListener("keydown", handleKeyDown);
      window.removeEventListener("scroll", show, true);
      window.removeEventListener("resize", show);
    };
  }, [position, show, hide]);

  return (
    <span
      ref={wrapperRef}
      className="contents"
      onPointerEnter={(event) => {
        if (event.pointerType === "mouse") show();
      }}
      onPointerLeave={(event) => {
        if (event.pointerType === "mouse") hide();
      }}
      onFocus={show}
      onBlur={hide}
      // スマホ等ホバーのない端末では、タップで表示する
      onClick={show}
    >
      {cloneElement(children, { "aria-describedby": id })}
      {/* 読み上げ用。表示中かどうかに関わらず aria-describedby から参照できるよう常に置く */}
      <span id={id} className="sr-only">
        {content}
      </span>
      {position
        ? createPortal(
            <span
              ref={tooltipRef}
              role="presentation"
              style={{ top: position.top, left: position.left }}
              className={
                "pointer-events-none fixed z-[60] w-max max-w-[min(18rem,calc(100vw-1rem))] -translate-x-1/2 rounded-md bg-foreground px-2.5 py-1.5 text-xs font-medium text-background shadow-md " +
                (position.placement === "top" ? "-translate-y-full" : "")
              }
            >
              {content}
            </span>,
            document.body,
          )
        : null}
    </span>
  );
}
