import React, { createContext, PropsWithChildren, useContext, useEffect, useState } from "react"

import {
    TokenSource
} from "../../vince";
import { useLocalStorage } from "../LocalStorageProvider";


type Props = {}

type ContextProps = {
    tokenSource: TokenSource | undefined
}

const defaultValues = {
    tokenSource: undefined,
}

export const TokenContext = createContext<ContextProps>(defaultValues)

export const TokenSourceProvider = ({ children }: PropsWithChildren<Props>) => {
    const { authPayload } = useLocalStorage()
    const [tokenSource, setTokenSource] = useState<TokenSource | undefined>(undefined)
    useEffect(() => {
        if (authPayload !== "") {
            setTokenSource(new TokenSource(authPayload))
        }
    }, [authPayload, setTokenSource])
    return (
        <TokenContext.Provider
            value={{ tokenSource }}
        >
            {children}
        </TokenContext.Provider>
    )
}

export const useTokenSource = () => {
    return useContext(TokenContext)
}