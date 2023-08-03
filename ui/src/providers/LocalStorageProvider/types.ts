export enum StoreKey {
    AUTH_PAYLOAD = "AUTH_PAYLOAD",
}

export const getValue = (key: StoreKey) => localStorage.getItem(key) ?? ""

export const setValue = (key: StoreKey, value: string) =>
    localStorage.setItem(key, value)

export type SettingsType = string | boolean | number


export type LocalConfig = {
    authPayload: string
}