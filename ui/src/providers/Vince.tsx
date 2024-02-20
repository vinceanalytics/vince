import { createContext, PropsWithChildren, useCallback, useContext, useEffect, useState } from "react";
import { Vince, Site } from "../vince";

type Props = {}

type ContextProps = {
    vince: Vince
    active: string,
    interval: Interval,
    sites: Site[]
    selectSite: (site: string) => void,
    setInterval: (i: Interval) => void,
    setVince: (v: Vince) => void,
}

export enum Interval {
    DATE = "date",
    MINUTE = "minute",
    HOUR = "hour",
    WEEK = "week",
    MONTH = "month"
}


const defaultValues = {
    vince: new Vince(),
    interval: Interval.DATE,
    active: "(no domain)",
    sites: [],
    selectSite: () => { },
    setInterval: () => { },
    setVince: () => { },
}


export const VinceContext = createContext<ContextProps>(defaultValues);


export const VinceProvider = ({ children }: PropsWithChildren<Props>) => {
    const [vince, setVince] = useState<Vince>(defaultValues.vince);
    const [sites, setSites] = useState<Site[]>(defaultValues.sites)
    const [active, selectSite] = useState<string>(defaultValues.active)
    const [interval, setInterval] = useState<Interval>(defaultValues.interval)
    useEffect(() => {
        vince.domains().then(({ domains }) => {
            if (domains && domains.length > 0) {
                setSites(domains)
                selectSite(domains[0].name)
            }
        }).catch(console.log)
    }, [vince, setSites, selectSite])


    return (
        <VinceContext.Provider value={{ vince, sites, active, setVince, selectSite, interval, setInterval }}>
            {children}
        </VinceContext.Provider>
    )
}

export const useVince = () => {
    return useContext(VinceContext)
}