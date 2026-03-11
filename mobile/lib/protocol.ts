export const PROTOCOL_VERSION = 1;

export type MessageType =
  | 'auth'
  | 'session_list'
  | 'session_event'
  | 'user_input'
  | 'approval_response'
  | 'ping'
  | 'pong'
  | 'machine_status';

export interface Envelope {
  v: number;
  type: MessageType;
  machine_id: string;
  session_id: string;
  timestamp: number;
  payload: string; // base64 encrypted blob
}

export interface Payload {
  kind: string;
  seq: number;
  data: any;
}

export interface Session {
  id: string;
  tool: string;
  project: string;
  status: string;
  started_at: string;
}

export type PayloadKind =
  | 'assistant_message'
  | 'tool_use'
  | 'tool_result'
  | 'user_message'
  | 'approval'
  | 'error';

export interface MachineStatus {
  machine_id: string;
  status: 'online' | 'offline' | 'active';
}
