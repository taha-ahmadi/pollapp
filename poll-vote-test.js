import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

const voteRates = {
  success: new Rate('vote_success_rate'),
};

const voteTrend = new Trend('vote_response_time');

const BASE_URL = __ENV.API_URL || 'http://localhost:8080';

const POLL_IDS = [1, 2];

const FIXED_OPTION_INDEX = 1;

export const options = {
  stages: [
    { duration: '20s', target: 10 },
    { duration: '30s', target: 20 },
    { duration: '1m', target: 20 },
    { duration: '30s', target: 50 },
    { duration: '1m', target: 50 },
    { duration: '30s', target: 0 },
  ],
  thresholds: {
    'vote_response_time': ['p(95)<500'],
    'vote_success_rate': ['rate>0.95'],
  },
};

export function setup() {
  const pollInfo = {};
  
  for (const pollId of POLL_IDS) {
    const fetchResponse = http.get(`${BASE_URL}/polls/${pollId}`);
    
    if (fetchResponse.status === 200) {
      try {
        const poll = JSON.parse(fetchResponse.body);
        pollInfo[pollId] = {
          id: pollId,
          optionsCount: poll.options.length,
          options: poll.options,
        };
        console.log(`Poll ${pollId} has ${poll.options.length} options`);
      } catch (e) {
        console.error(`Error parsing poll ${pollId} response: ${e}`);
        pollInfo[pollId] = {
          id: pollId,
          optionsCount: 3,
          options: ["Option 1", "Option 2", "Option 3"],
        };
      }
    } else {
      console.error(`Failed to fetch poll ${pollId}: ${fetchResponse.status} ${fetchResponse.body}`);
      pollInfo[pollId] = {
        id: pollId,
        optionsCount: 3,
        options: ["Option 1", "Option 2", "Option 3"],
      };
    }
  }
  
  return { pollInfo };
}

export default function(data) {
  const userId = (__VU * 1000) + __ITER; 
  
  const pollIdIndex = __ITER % POLL_IDS.length;
  const pollId = POLL_IDS[pollIdIndex];
  
  const pollInfo = data.pollInfo[pollId];
  if (!pollInfo) {
    console.error(`No information available for poll ${pollId}, skipping`);
    return;
  }
  
  const optionIndex = FIXED_OPTION_INDEX;
  
  console.log(`User ${userId} voting on poll ${pollId}, option ${optionIndex} (${pollInfo.options[optionIndex] || 'Unknown'})`);
  
  const voteRequest = {
    userId: userId,
    optionIndex: optionIndex
  };
  
  const voteStart = new Date();
  const voteResp = http.post(
    `${BASE_URL}/polls/${pollId}/vote`,
    JSON.stringify(voteRequest),
    {
      headers: { 'Content-Type': 'application/json' },
      timeout: '3s'
    }
  );
  voteTrend.add(new Date() - voteStart);
  
  const voteSuccess = check(voteResp, {
    'vote_status_is_200': (r) => r.status === 200 || r.status === 204,
    'not_server_error': (r) => r.status < 500,
  });
  voteRates.success.add(voteSuccess);
  
  if (voteResp.status !== 200 && voteResp.status !== 204) {
    console.error(`Vote failed for pollId: ${pollId}, userId: ${userId}, optionIndex: ${optionIndex}, statusCode: ${voteResp.status}, body: ${voteResp.body}`);
  }
  
  sleep(1);
} 