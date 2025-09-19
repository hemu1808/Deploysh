
import React from 'react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import type { Metric } from '../types';

interface MetricChartProps {
  title: string;
  data: Metric[];
  dataKey: string;
  color: string;
}

const CustomTooltip: React.FC<any> = ({ active, payload, label }) => {
  if (active && payload && payload.length) {
    return (
      <div className="bg-gray-700 p-2 border border-gray-600 rounded shadow-lg">
        <p className="label text-gray-300">{`${new Date(label).toLocaleTimeString()}`}</p>
        <p className="intro text-sm" style={{ color: payload[0].color }}>{`${payload[0].name} : ${payload[0].value.toFixed(2)}`}</p>
      </div>
    );
  }
  return null;
};

const MetricChart: React.FC<MetricChartProps> = ({ title, data, dataKey, color }) => {
  return (
    <div className="bg-gray-900/50 p-4 rounded-lg">
      <h4 className="text-md font-semibold text-gray-300 mb-4">{title}</h4>
      <div style={{ width: '100%', height: 250 }}>
          <ResponsiveContainer>
              <LineChart
                  data={data}
                  margin={{ top: 5, right: 20, left: -10, bottom: 5 }}
              >
                  <CartesianGrid strokeDasharray="3 3" stroke="#4a5568" />
                  <XAxis 
                      dataKey="time" 
                      stroke="#a0aec0" 
                      tickFormatter={(timeStr) => new Date(timeStr).toLocaleTimeString()} 
                      fontSize={12}
                  />
                  <YAxis stroke="#a0aec0" fontSize={12} />
                  <Tooltip content={<CustomTooltip />} />
                  <Line type="monotone" dataKey={dataKey} name={title.split(' ')[0]} stroke={color} strokeWidth={2} dot={false} />
              </LineChart>
          </ResponsiveContainer>
      </div>
    </div>
  );
};

export default MetricChart;
