export enum StoreKey {
    AUTH_PAYLOAD = "AUTH_PAYLOAD",
    QUERY_TEXT = "query.text",
    EDITOR_LINE = "editor.line",
    EDITOR_COL = "editor.col",
    NOTIFICATION_ENABLED = "notification.enabled",
    NOTIFICATION_DELAY = "notification.delay",
    EDITOR_SPLITTER_BASIS = "splitter.editor.basis",
    RESULTS_SPLITTER_BASIS = "splitter.results.basis",
}

export const getValue = (key: StoreKey) => localStorage.getItem(key) ?? ""

export const setValue = (key: StoreKey, value: string) =>
    localStorage.setItem(key, value)

export type SettingsType = string | boolean | number


export type LocalConfig = {
    authPayload: string
    editorSplitterBasis: number
    resultsSplitterBasis: number
}

export const parseBoolean = (value: string, defaultValue: boolean): boolean =>
    value ? value === "true" : defaultValue

export const parseInteger = (value: string, defaultValue: number): number =>
    isNaN(parseInt(value)) ? defaultValue : parseInt(value)