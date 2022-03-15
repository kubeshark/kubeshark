import { useState, useRef } from "react";

enum WebSocketReadyState {
  CONNECTING,
  OPEN,
  CLOSING,
  CLOSED
}

export const DEFAULT_QUERY = "leftOff(-1)";

export default function useWS(wsUrl: string) {
  const [message, setMessage] = useState(null);
  const [error, setError] = useState(null);
  const [isOpen, setisOpen] = useState(false);
  const ws = useRef(null);

  const onMessage = (e) => { setMessage(e) }
  const onError = (e) => setError(e)
  const onOpen = () => { setisOpen(true) }
  const onClose = () => setisOpen(false)

  const openSocket = () => {
    ws.current = new WebSocket(wsUrl)
    ws.current.addEventListener("message", onMessage)
    ws.current.addEventListener("error", onError)
    ws.current.addEventListener("open", onOpen)
    ws.current.addEventListener("close", onClose)
  }

  const closeSocket = () => {
    ws.current.readyState === WebSocketReadyState.OPEN && ws.current.close();
    ws.current.removeEventListener("message", onMessage)
    ws.current.removeEventListener("error", onError)
    ws.current.removeEventListener("open", onOpen)
    ws.current.removeEventListener("close", onClose)
  }

  const sendQuery = (query: string) => {
    if (ws.current && (ws.current.readyState === WebSocketReadyState.OPEN)) {
      ws.current.send(JSON.stringify({ "query": query, "enableFullEntries": false }));
    }
  }

  return { message, error, isOpen, openSocket, closeSocket, sendQuery }
}