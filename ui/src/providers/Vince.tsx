import { createContext, PropsWithChildren, useCallback, useContext, useEffect, useState } from "react";
import { Vince, Site } from "../vince";

type Props = {}

type ContextProps = {
    vince: Vince
    active: string,
    sites: Site[]
    selectSite: (site: string) => void,
}

const defaultValues = {
    vince: new Vince(),
    active: "(no domain)",
    sites: [],
    selectSite: () => { },
}

interface Sites {
    active: string,
    domains: Site[]
}

export const VinceContext = createContext<ContextProps>(defaultValues);


export const VinceProvider = ({ children }: PropsWithChildren<Props>) => {
    const [vince, setVince] = useState<Vince>(defaultValues.vince);
    const [sites, setSites] = useState<Site[]>(defaultValues.sites)
    const [active, setSite] = useState<string>(defaultValues.active)
    useEffect(() => {
        vince.domains().then(({ domains }) => {
            if (domains && domains.length > 0) {
                console.log(domains)
                setSites(domains)
                setSite(domains[0].name)
            }
        }).catch(console.log)
    }, [vince])

    const selectSite = useCallback((active: string) => {
        setSite(active)
    }, [])


    return (
        <VinceContext.Provider value={{ vince, sites, active, selectSite }}>
            {children}
        </VinceContext.Provider>
    )
}

export const useVince = () => {
    return useContext(VinceContext)
}