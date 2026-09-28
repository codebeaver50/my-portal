"use client";

import { useState } from "react";
import type { FormEvent } from "react";
import { Button, Input, Modal, Textarea, useToast } from "@/components/ui";
import { sendContactMessage } from "../_lib/api";
import { contactSchema } from "../_lib/schema";
import type { ContactFormData } from "../_lib/schema";

export function ContactForm() {
  const { showToast } = useToast();
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [message, setMessage] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [isPending, setIsPending] = useState(false);
  const [isSent, setIsSent] = useState(false);
  const [confirmData, setConfirmData] = useState<ContactFormData | null>(null);

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError(null);

    const parsed = contactSchema.safeParse({ name, email, message });
    if (!parsed.success) {
      setError(parsed.error.issues[0]?.message ?? "入力内容を確認してください");
      return;
    }

    setConfirmData(parsed.data);
  }

  function handleCloseConfirm() {
    if (isPending) return;
    setConfirmData(null);
  }

  async function handleSend() {
    if (!confirmData) return;

    setIsPending(true);
    try {
      await sendContactMessage(confirmData);
      setConfirmData(null);
      setIsSent(true);
      showToast("お問い合わせを送信しました");
    } catch (err) {
      showToast(err instanceof Error ? err.message : "送信に失敗しました", "error");
    } finally {
      setIsPending(false);
    }
  }

  function handleReset() {
    setName("");
    setEmail("");
    setMessage("");
    setIsSent(false);
  }

  return (
    <>
      <form className="flex flex-col gap-4" onSubmit={handleSubmit}>
        {isSent ? (
          <p className="rounded-lg border border-success/30 bg-success-bg px-4 py-3 text-sm text-success">
            送信しました。内容は下記の通りです。
          </p>
        ) : null}
        <Input
          id="contact-name"
          label="お名前"
          value={name}
          onChange={(event) => setName(event.target.value)}
          disabled={isPending || isSent}
          required
        />
        <Input
          id="contact-email"
          label="メールアドレス"
          type="email"
          value={email}
          onChange={(event) => setEmail(event.target.value)}
          disabled={isPending || isSent}
          required
        />
        <Textarea
          id="contact-message"
          label="お問い合わせ内容"
          rows={6}
          value={message}
          onChange={(event) => setMessage(event.target.value)}
          disabled={isPending || isSent}
          required
        />
        {error ? <p className="text-sm text-danger">{error}</p> : null}
        {isSent ? (
          <Button type="button" variant="secondary" className="w-auto" onClick={handleReset}>
            別の内容を送る
          </Button>
        ) : (
          <Button type="submit" variant="accent" className="w-auto" disabled={isPending}>
            確認する
          </Button>
        )}
      </form>

      <Modal
        isOpen={confirmData !== null}
        onClose={handleCloseConfirm}
        title="送信内容の確認"
        footer={
          <>
            <Button
              type="button"
              variant="secondary"
              className="w-auto"
              onClick={handleCloseConfirm}
              disabled={isPending}
            >
              キャンセル
            </Button>
            <Button
              type="button"
              variant="accent"
              className="w-auto"
              onClick={handleSend}
              disabled={isPending}
            >
              {isPending ? "送信中..." : "送信する"}
            </Button>
          </>
        }
      >
        <p className="text-sm">この内容で送信してもよろしいですか？</p>
      </Modal>
    </>
  );
}
