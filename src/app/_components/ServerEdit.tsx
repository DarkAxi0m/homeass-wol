"use client";
import React, { useState } from "react";
import { Actions, ActionsByType, Server } from "~/lib/server";

interface ServerEditDialogProps {
  server: Server | null;
  onSave: (server: Server) => void;
  onCancel: () => void;
}

const ServerEditDialog: React.FC<ServerEditDialogProps> = ({
  server,
  onSave,
  onCancel,
}) => {
  const [name, setName] = useState(server?.name || "");
  const [startType, setStartType] = useState(server?.startType || "");
  const [stopType, setStopType] = useState(server?.stopType || "");
  const [checkType, setCheckType] = useState(server?.checkType || "");
  const [reqValues, setReqValues] =
    useState<Record<string, Action>>(geteReqValues());
  const [settings, setSettings] = useState<Record<string, string>>(
    server?.settings || {},
  );

  const handleSave = () => {
    if (server) {
      onSave({ ...server, name, startType, stopType, checkType, settings });
    }
  };

  function geteReqValues() {
    var s = {};
    const StartAction = Actions[startType.toLowerCase()];
    if (StartAction) {
      s = { ...s, ...StartAction.reqValues };
    }

    const StopAction = Actions[stopType.toLowerCase()];
    if (StopAction) {
      s = { ...s, ...StopAction.reqValues };
    }

    const CheckAction = Actions[checkType.toLowerCase()];
    if (CheckAction) {
      s = { ...s, ...CheckAction.reqValues };
    }

    return s;
  }

  React.useEffect(() => {
    setReqValues(geteReqValues());
  }, [startType, stopType, checkType]);

  function updateSettings(area: string, value: string) {
    let s = { ...settings };
    s[area] = value;
    console.log(area, value, s);
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
                defaultValue={startType}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg"
                onChange={(e) => setStartType(e.target.value)}
              >
                <option value="">No Action</option>
                {ActionsByType("start").map((v, i) => (
                  <option key={i}>{v.name}</option>
                ))}
              </select>
            </div>
            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Stop Type
              </label>

              <select
                defaultValue={stopType}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg"
                onChange={(e) => setStopType(e.target.value)}
              >
                <option value="">No Action</option>
                {ActionsByType("stop").map((v, i) => (
                  <option key={i}>{v.name}</option>
                ))}
              </select>
            </div>
            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Check Type
              </label>

              <select
                defaultValue={checkType}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg"
                onChange={(e) => setCheckType(e.target.value)}
              >
                <option value="">No Action</option>
                {ActionsByType("check").map((v, i) => (
                  <option key={i}>{v.name}</option>
                ))}
              </select>
            </div>{" "}
          </div>
          <div className="p-2">
            {Object.keys(reqValues).map((s) => (
              <div key={s} className="mb-4">
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  {s} <small>{reqValues[s]}</small>
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
