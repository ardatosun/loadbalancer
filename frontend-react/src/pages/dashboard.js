import DataTable from "./components/DataTable"
import Searchbar from "./components/Searchbar"
import Sort from "./components/Sort"
import { GiHealthNormal } from "react-icons/gi";
import { CgDanger } from "react-icons/cg";
import { SlGraph } from "react-icons/sl";
import { useEffect, useState } from "react";    
import healthStatus from "../enums/healthStatus.json";

export default function Dashboard() {

    const { backends } = healthStatus;
    const [backend, setBackend] = useState([]);
    const [healthyIP, setHealthyIP] = useState([]);
    const [unhealthyIP, setUnhealthyIP] = useState([]);

    useEffect(()=>{
       if(backends.length > 0){
        setBackend(backends);
       }
    },[backends]);

    useEffect(()=>{
        setHealthyIP(prevState => {
            const healthyServer = backend.filter(ip => ip.status === "healthy");
            return healthyServer;
        })
        setUnhealthyIP(prevState => {
            const unhealthyServer = backend.filter(ip => ip.status === "unhealthy");
            return unhealthyServer;
        })
    },[backend])

    return (
        <div className="bg-grey h-full p-4">
            <div className="header flex p-2 border-b-2 border-lightgrey gap-3">
                <Searchbar data={backends} updateData={setBackend}/>
                <div className="border-r-2 border-lightgrey"></div>
                <div className="w-1/12 flex items-center justify-center">
                    <Sort data={backends} updateData={setBackend}/>
                </div>
            </div>
            <div className="mid p-2 flex flex-col gap-3">
                <span className="tracking-wider text-xl">Statistics</span>
                <div className="w-full flex gap-5">
                    <div className="w-3/12 border-1 border-lightgrey flex flex-col p-5 gap-2 rounded-xl">
                        <div className="flex items-center justify-between">
                            <div className="flex flex-col gap-1">
                                <span className="tracking-wider text-base">Healthy</span>
                                <span className="text-lg tracking-wider">{`${healthyIP.length} out of ${backend.length}`}</span>
                            </div>
                            <div className="bg-darkgreen h-12 w-12 flex items-center justify-center rounded-full"><GiHealthNormal color="#62d5a4" size={20}/></div>
                        </div>
                        
                        <div className="h-3 w-12/12 bg-darkgreen rounded-md">
                            <div 
                                className={`h-3 bg-limegreen rounded-md`}
                                style={{ width: `${(healthyIP.length / backend.length) * 100}%` }}
                            >
                                &nbsp;
                            </div>
                        </div>
                    </div>
                    <div className="w-3/12 border-1 border-lightgrey flex flex-col p-5 gap-2 rounded-xl">
                        <div className="flex items-center justify-between">
                            <div className="flex flex-col gap-1">
                                <span className="tracking-wider text-base">Unhealthy</span>
                                <span className="text-lg tracking-wider">{`${unhealthyIP.length} out of ${backend.length}`}</span>
                            </div>
                            <div className="bg-darkred h-12 w-12 flex items-center justify-center rounded-full"><CgDanger color="#ff3333" size={25}/></div>
                        </div>
                        <div className="h-3 w-12/12 bg-darkred rounded-md">
                            <div 
                                className={`h-3 bg-lightred rounded-md`}
                                style={{ width: `${(unhealthyIP.length / backend.length) * 100}%` }}
                            >
                                &nbsp;
                            </div>
                        </div>
                    </div>
                    <div className="w-6/12 border-1 border-lightgrey flex p-5 gap-4 items-center rounded-xl">
                        <div>
                            <div className="bg-lightgrey h-16 w-16 flex items-center justify-center rounded-full"><SlGraph size={35}/></div>
                        </div>
                        <div className="flex flex-col gap-1">
                            <span className="tracking-wider text-lg">{`Total IP Address : ${backend.length}`}</span>
                            <span className="text-base tracking-wider text-whitegrey">This includes dynamic and static IP addresses assigned within the network.</span>
                        </div>
                    </div>
                </div>
            </div>
            <div className="bottom p-2 gap-2 flex flex-col">
                <span className="text-xl tracking-wider">IP Address List</span>
                <DataTable data={backend}/>
            </div>
        </div>
    )
  }