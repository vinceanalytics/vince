import { createContext, PropsWithChildren, useContext, useState } from "react";
import { Vince } from "../vince";

type Props = {}

type ContextProps = {
    vince: Vince
}

const defaultValues = {
    vince: new Vince(),
}

export const VinceContext = createContext<ContextProps>(defaultValues);


export const VinceProvider = ({ children }: PropsWithChildren<Props>) => {
    const [vince, setVince] = useState<Vince>(defaultValues.vince);
    return (
        <VinceContext.Provider value={{ vince }}>
            {children}
        </VinceContext.Provider>
    )
}

export const useVince = () => {
    return useContext(VinceContext)
}