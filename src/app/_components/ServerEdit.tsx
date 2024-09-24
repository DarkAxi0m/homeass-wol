"use client";
import { Server, Action } from "@prisma/client";
import React, { useState } from "react";
import { api } from "~/trpc/react";

interface ServerEditDialogProps {
  server: Server;
  actions: Action[];
  onSave: (server: Server) => void;
  onCancel: () => void;
}

type SettingRec = Record<string,string>

const ServerEditDialog: React.FC<ServerEditDialogProps> =  ({
  server,
  actions,
  onSave,
  onCancel,
}) => {
 
  const [name, setName] = useState(server?.name || "");
  const [startTypeId, setStartTypeId] = useState(server.startTypeId);
  const [stopTypeId, setStopTypeId] = useState(server.stopTypeId);
  const [checkTypeId, setCheckTypeId] = useState(server.checkTypeId );
  const [reqValues, setReqValues] =
    useState<Record<string, Action>>(geteReqValues());
  const [settings, setSettings] = useState<SettingRec>(
    JSON.parse( server?.settings || "{}") as SettingRec,
  );

  const handleSave = () => {
    if (server) {
        console.log('Handle Save', server)
      onSave({ ...server, name, startTypeId, stopTypeId, checkTypeId, settings:JSON.stringify( settings ) });
    }
  };

  function geteReqValues() {
    var s = {};

    if (!actions) return s;

    const a = actions.find((a)=>a.id == startTypeId)
    if (a) {
       s = {...s, ... JSON.parse(a.reqValues)}
    }
     const b  = actions.find((a)=>a.id == stopTypeId)
    if (b) {
       s = {...s, ... JSON.parse(b.reqValues)}
    }
     const c = actions.find((a)=>a.id == checkTypeId)
    if (c) {
       s = {...s, ... JSON.parse(c.reqValues)}
    }
 
    return s;
  }

  React.useEffect(() => {
    setReqValues(geteReqValues());
  }, [startTypeId, stopTypeId, checkTypeId]);

  function updateSettings(area: string, value: string) {
    let s = { ...settings };
    s[area] = value;
    setSettings(s);
  }

  if (!server) return null;

  return (
    <div className="fixed inset-0 bg-gray-600 bg-opacity-50 flex justify-center items-center">
      <div className="bg-white p-6 rounded-lg shadow-lg min-w-w-96">
        <h2 className="text-xl font-bold mb-4">Edit Server</h2>
        <div className="grid grid-cols-2">
          <div className="p-2">
            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Name
              </label>
              <input
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg"
              />
            </div>
            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Start Type
              </label>

              <select
                defaultValue={startTypeId ?? ""}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg"
                onChange={(e) => setStartTypeId(parseInt( e.target.value))}
              >
                <option value="0">No Action</option>
                {actions && actions.filter(a=>a.type=="start").map(a=><option value={a.id.toString()}>{a.name}</option>)}
              </select>
            </div>
            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Stop Type
              </label>

              <select
                defaultValue={stopTypeId ?? ""}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg"
                onChange={(e) => setStopTypeId(parseInt(e.target.value))}
              >
                <option value="0">No Action</option>
                {actions && actions.filter(a=>a.type=="stop").map(a=><option value={a.id.toString()}>{a.name}</option>)}
              </select>
            </div>
            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Check Type
              </label>

              <select
                defaultValue={checkTypeId ?? ""}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg"
                onChange={(e) => setCheckTypeId(parseInt(e.target.value))}
              >
                <option value="0">No Action</option>
                {actions && actions.filter(a=>a.type=="check").map(a=><option value={a.id.toString()}>{a.name}</option>)}
              </select>
            </div>{" "}
          </div>
          <div className="p-2">
            {Object.keys(reqValues).map((s) => (
              <div key={s} className="mb-4">
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  {s} <small></small>
                </label>
                <input
                  type="text"
                  value={settings[s]}
                  onChange={(e) => updateSettings(s, e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg"
                />
              </div>
            ))}
          </div>
        </div>
        <div className="flex justify-end">
          <button
            onClick={onCancel}
            className="mr-2 px-4 py-2 text-gray-700 bg-gray-200 rounded-lg hover:bg-gray-300"
          >
            Cancel
          </button>
          <button
            onClick={handleSave}
            className="px-4 py-2 text-white bg-blue-600 rounded-lg hover:bg-blue-700"
          >
            Save
          </button>
        </div>
      </div>
    </div>
  );
};

export default ServerEditDialog;
