export interface ServiceInfo {
  name: string;
  fullName: string;
  methods: MethodInfo[];
}

export interface MethodInfo {
  name: string;
  fullName: string;
  inputType: string;
  outputType: string;
  clientStreaming: boolean;
  serverStreaming: boolean;
}

export interface FieldInfo {
  name: string;
  type: string;
  repeated: boolean;
  optional: boolean;
  description?: string;
}

export interface MessageSchema {
  name: string;
  fields: FieldInfo[];
}
