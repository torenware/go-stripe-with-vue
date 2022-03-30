// Types for the Base Forms components

export enum FieldSetState {
    REQUIRED_MISSING = -1
}

export type JSPO = {
    [key: string]: string | number
}

export type ProcessSubmitFunc = (obj: JSPO, form: HTMLFormElement | null) => void;
export type GatherValueFunc = (form: HTMLFormElement) => JSPO;
