import { useState, useRef } from "react";

enum WebSocketReadyState {
  CONNECTING,
  OPEN,
  CLOSING,
  CLOSED
}

export const DEFAULT_LEFTOFF = `latest`;
export const DEFAULT_FETCH = 50;
export const DEFAULT_FETCH_TIMEOUT_MS = 3000;

const useWS = (wsUrl: string) => {
  const [message, setMessage] = useState(null);
  const [error, setError] = useState(null);
  const [isOpen, setisOpen] = useState(false);
  const ws = useRef(null);

  const onMessage = (e) => { setMessage(e) }
  const onError = (e) => setError(e)
  const onOpen = () => { setisOpen(true) }
  const onClose = () => {
    setisOpen(false)
  }

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

  const sendQueryWhenWsOpen = (query) => {
    setTimeout(() => {
      if (ws?.current?.readyState === WebSocket.OPEN) {
        ws.current.send(JSON.stringify({"query": query, "enableFullEntries": false}));
      } else {
        sendQueryWhenWsOpen(query);
      }
    }, 500)
  }

  return { message, error, isOpen, openSocket, closeSocket, sendQueryWhenWsOpen }
}

export default useWS
