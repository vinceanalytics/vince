import React, { createContext, PropsWithChildren } from "react"

import { Client } from "../../vince";

const client = new Client();

type Props = {}

type ContextProps = {
    vince: Client
}

const defaultValues = {
    vince: client,
}

export const VinceContext = createContext<ContextProps>(defaultValues)

export const VinceProvider = ({ children }: PropsWithChildren<Props>) => {
    return (
        <VinceContext.Provider
            value={{
                vince: client
            }}
        >
            {children}
        </VinceContext.Provider>
    )
}