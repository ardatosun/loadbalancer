import React, { useEffect, useState } from "react";
import { TbArrowsDownUp } from "react-icons/tb";

export default function Chips({bgColor, txtColor, ledColor, text}) {

    return (
        <div className={`rounded-lg flex justify-center items-center w-5/12 h-8 px-2`}>
            <div className={` h-10 w-3/12 flex items-center `}>
                <div className={`${ledColor} rounded-full h-3 w-3`}>&nbsp;</div>
            </div>
            <span className={`w-9/12 tracking-wider text-base flex items-center capitalize`}>{text}</span>
        </div>
    )
}