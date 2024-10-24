import Link from "next/link";
import { useRouter } from "next/router";
import React, { useEffect, useState } from "react";
import { MdOutlineSpaceDashboard, MdOutlineHome } from "react-icons/md";

export default function Sidebar() {

    const router = useRouter();
    const [path, setPath] = useState(router.pathname);

    const navigateToHome = () => {
        router.push('/');
        setPath("/");
    };

    const navigateToDashboard = () => {
        router.push('/dashboard');
        setPath("/dashboard");
    };

    return (
        <div className="w-2/12 p-6 h-full bg-black">
            <div className="py-6">
                <span className="text-xl tracking-wider">Server Status</span>
            </div>
            <div className="flex flex-col">
                <span className="text-sm font-bold tracking-wider">General</span>
                <div className="flex flex-col gap-2 py-2">
                    {/* <div 
                        className={`text-s flex gap-1.5 px-4 py-1 rounded-md cursor-pointer ${path === "/" ? "bg-grey" : null}`}
                        onClick={navigateToHome}
                    >
                        <div className="w-6 flex items-center"><MdOutlineHome color="#62d5a4" size={21}/></div>
                        <span className="tracking-wider">Home</span>
                    </div> */}
                    <div 
                        className={`text-s flex gap-1.5 px-4 py-1 rounded-md cursor-pointer ${path === "/dashboard" ? "bg-grey" : null}`}
                        onClick={navigateToDashboard}
                    >
                        <div className="w-6 flex items-center"><MdOutlineSpaceDashboard color="#62d5a4" size={18}/></div>
                        <span className="tracking-wider">Dashboard</span>
                    </div>
                </div>
               
            </div>
        </div>
    )
}