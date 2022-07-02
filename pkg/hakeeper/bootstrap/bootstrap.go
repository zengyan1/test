// Copyright 2021 - 2022 Matrix Origin
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bootstrap

import (
	"errors"
	"github.com/matrixorigin/matrixone/pkg/hakeeper/checkers/util"
	pb "github.com/matrixorigin/matrixone/pkg/pb/logservice"
	"github.com/matrixorigin/matrixone/pkg/pb/metadata"
	"sort"
)

type Manager struct {
	cluster pb.ClusterInfo
}

func NewBootstrapManager(cluster pb.ClusterInfo) *Manager {
	var dnShard []metadata.DNShardRecord
	var logShard []metadata.LogShardRecord

	if cluster.DNShards != nil {
		dnShard = make([]metadata.DNShardRecord, len(cluster.DNShards))
		copy(dnShard, cluster.DNShards)
	}

	if cluster.LogShards != nil {
		logShard = make([]metadata.LogShardRecord, len(cluster.LogShards))
		copy(logShard, cluster.LogShards)
	}

	manager := &Manager{
		cluster: pb.ClusterInfo{
			DNShards:  dnShard,
			LogShards: logShard,
		},
	}

	return manager
}

func (bm *Manager) Bootstrap(alloc util.IDAllocator,
	dn pb.DNState, log pb.LogState) (commands []pb.ScheduleCommand, err error) {

	logStores := sortLogStoresByTick(log.Stores)
	dnStores := sortDNStoresByTick(dn.Stores)

	for _, shardRecord := range bm.cluster.LogShards {
		if shardRecord.NumberOfReplicas > uint64(len(logStores)) {
			return nil, errors.New("not enough log stores")
		}

		initialMembers := make(map[uint64]string)

		for _, storeInfo := range logStores {
			if uint64(len(initialMembers)) == shardRecord.NumberOfReplicas {
				break
			}

			replicaID, ok := alloc.Next()
			if !ok {
				return nil, errors.New("id allocator error")
			}

			initialMembers[replicaID] = storeInfo
		}

		for replicaID, uuid := range initialMembers {
			commands = append(commands,
				pb.ScheduleCommand{
					UUID: uuid,
					ConfigChange: &pb.ConfigChange{
						Replica: pb.Replica{
							UUID:      uuid,
							ShardID:   shardRecord.ShardID,
							ReplicaID: replicaID,
						},
						ChangeType:     pb.AddReplica,
						InitialMembers: initialMembers,
					},
					ServiceType: pb.LogService,
				})
		}
	}

	for _, dnRecord := range bm.cluster.DNShards {
		for _, uuid := range dnStores {
			replicaID, ok := alloc.Next()
			if !ok {
				return nil, errors.New("id allocator error")
			}

			commands = append(commands, pb.ScheduleCommand{
				UUID: uuid,
				ConfigChange: &pb.ConfigChange{
					Replica: pb.Replica{
						UUID:      uuid,
						ShardID:   dnRecord.ShardID,
						ReplicaID: replicaID,
					},
					ChangeType: pb.AddReplica,
				},
				ServiceType: pb.DnService,
			})
			break
		}
	}

	return
}

func (bm *Manager) CheckBootstrap(log pb.LogState) bool {
	for _, shardInfo := range log.Shards {
		var shardRecord metadata.LogShardRecord
		for _, record := range bm.cluster.LogShards {
			if record.ShardID == shardInfo.ShardID {
				shardRecord = record
				break
			}
		}
		if shardRecord.ShardID == 0 {
			return false
		}

		if uint64(len(shardInfo.Replicas))*2 <= shardRecord.NumberOfReplicas {
			return false
		}
	}

	return true
}

func sortLogStoresByTick(logStores map[string]pb.LogStoreInfo) []string {
	storeSlice := make([]struct {
		uuid string
		pb.LogStoreInfo
	}, 0, len(logStores))
	uuidSlice := make([]string, 0, len(logStores))

	for uuid, storeInfo := range logStores {
		storeSlice = append(storeSlice, struct {
			uuid string
			pb.LogStoreInfo
		}{uuid, storeInfo})
	}

	sort.Slice(storeSlice, func(i, j int) bool {
		return storeSlice[i].Tick > storeSlice[j].Tick
	})

	for _, store := range storeSlice {
		uuidSlice = append(uuidSlice, store.uuid)
	}

	return uuidSlice
}

func sortDNStoresByTick(dnStores map[string]pb.DNStoreInfo) []string {
	storeSlice := make([]struct {
		uuid string
		pb.DNStoreInfo
	}, 0, len(dnStores))
	uuidSlice := make([]string, 0, len(dnStores))

	for uuid, storeInfo := range dnStores {
		storeSlice = append(storeSlice, struct {
			uuid string
			pb.DNStoreInfo
		}{uuid, storeInfo})
	}

	sort.Slice(storeSlice, func(i, j int) bool {
		return storeSlice[i].Tick > storeSlice[j].Tick
	})

	for _, store := range storeSlice {
		uuidSlice = append(uuidSlice, store.uuid)
	}

	return uuidSlice
}