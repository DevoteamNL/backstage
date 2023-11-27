import { useApi } from '@backstage/core-plugin-api';
import { useContext, useEffect, useState } from 'react';
import { MetricData } from '../models/MetricData';
import { groupDataServiceApiRef } from '../services/GroupDataService';
import { MetricContext } from '../services/MetricContext';

export const useMetricData = (type: string) => {
  const groupDataService = useApi(groupDataServiceApiRef);
  const [chartData, setChartData] = useState<MetricData | undefined>();
  const [error, setError] = useState<Error | undefined>();
  const { aggregation, team, project } = useContext(MetricContext);

  useEffect(() => {
    groupDataService
      .retrieveMetricDataPoints({
        type: type,
        team: team,
        aggregation: aggregation,
        project: project,
      })
      .then(response => {
        if (response.dataPoints.length > 0) {
          setChartData(response);
        } else {
          setError(new Error('No data found'));
        }
      }, setError);
  }, [aggregation, team, project, groupDataService, type]);

  return { error: error, chartData: chartData };
};
