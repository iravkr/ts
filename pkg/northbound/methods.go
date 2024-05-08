package northbound

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/onosproject/cco-mon/pkg/mho"
	pb "github.com/onosproject/cco-mon/pkg/proto"
	ocnstorage "github.com/onosproject/cco-mon/pkg/store/store/ocn"
	"github.com/onosproject/onos-mho/pkg/store"
)

func (m *MonitoringServer) GetRsrpReports(_ *pb.NoParam, stream pb.CCOMonitoringService_GetRsrpReportsServer) error {

	now := time.Now()
	ues := m.manager.GetUEs(context.Background()) // please check config and uncommet for ocs..this value might be  
	keys := make([]string, 0, len(ues))
	for k := range ues {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	if len(ues) > 0 {
		for _, key := range keys {

			ueID := ues[key].UeID
			fiveQi := ues[key].FiveQi
			cgiString := ues[key].CGIString
			if cgiString == "" {
				cgiString = "NONE"
			}
			rsrpServing := ues[key].RsrpServing

			cgi_keys := make([]string, 0, len(ues[key].RsrpNeighbors))
			for k := range ues[key].RsrpNeighbors {
				cgi_keys = append(cgi_keys, k)
			}
			sort.Strings(cgi_keys)

			RsrpNeigborsMap := make(map[string]int32)
			for _, cgi := range cgi_keys {
				RsrpNeigborsMap[cgi] = ues[key].RsrpNeighbors[cgi]

			}

			res := &pb.RsrpInfo{
				Time:          now.Format("2006-01-02 15:04:05"),
				Fiveqi:        fiveQi,
				Ueid:          ueID,
				Cgi:           cgiString,
				RsrpServing:   rsrpServing,
				RsrpNeighbors: RsrpNeigborsMap,
			}
			log.Infof("Time %v, UeID %v, S_CGI %v, RsrpServing %v, RsrpNeighbors %v", now.Format("2006-01-02 15:04:05"), ueID, cgiString, rsrpServing, RsrpNeigborsMap)

			if err := stream.Send(res); err != nil {
				return err

			}

		}

	}

	return nil
}

func (m *Manager) GetUEs(ctx context.Context) map[string]mho.UeData {
	m.mutex.Lock() // locking the synch access to shared resouce to ensure only this goroutine can access the preotected resource at a time
	defer m.mutex.Unlock()
	output := make(map[string]mho.UeData)
	chEntries := make(chan *store.Entry, 1024) // creating a buffered channel capable of holding up to 1024 pointers to store.Entry objects
	err := m.ueStore.Entries(ctx, chEntries)
	if err != nil {
		log.Warn(err)
		return output
	}
	for entry := range chEntries {
		ueData := entry.Value.(mho.UeData)
		output[ueData.UeID] = ueData
	}
	return output
}

func (m *MonitoringServer) GetCellInfo(_ *pb.NoParam, stream pb.CCOMonitoringService_GetCellInfoServer) error {
	cellInfoMap := make(map[string]float32)
	cellInfo := m.manager.metricStore.GetAllEntries(context.Background())

	for _, value := range cellInfo {
		cellInfoMap[value.Value.CGIstring] = float32(value.Value.PTX)
	}

	for value, key := range cellInfoMap {
		res := &pb.CellInfo{
			Cgi: value,
			Ptx: key,
		}

		if err := stream.Send(res); err != nil {
			return err
		}

	}
	return nil

}

func (m *MonitoringServer) SetCellPTX(ctx context.Context, cellInfo *pb.CellInfo) (*pb.Response, error) {

	entries := m.manager.metricStore.GetAllEntries(ctx)

	for key, value := range entries {
		if value.Value.CGIstring == cellInfo.Cgi {
			err := m.manager.metricStore.UpdatePtx(ctx, key, int32(cellInfo.Ptx))
			if err != nil {
				return &pb.Response{
					Response: "Error While Updating Cell Info",
				}, err
			}
		}
	}

	return &pb.Response{
		Response: "Updated",
	}, nil
}

// GetOcn gets Ocn map
func (m *MonitoringServer) GetOcn(ctx context.Context, _ *pb.GetOcnRequest) (*pb.GetOcnResponse, error) {
	ch := make(chan ocnstorage.Entry)
	go func(ch chan ocnstorage.Entry) {
		err := m.manager.ocnStore.ListAllInnerElement(ctx, ch) // this is empty?
		if err != nil {
			log.Warn(err)
			close(ch)
		}
	}(ch)

	mapOcnResp := make(map[string]*pb.OcnRecord)

	// Init map in ocnresp message
	for e := range ch {
		key := fmt.Sprintf("%s:%s:%s:%s", e.Key.NodeID, e.Key.PlmnID, e.Key.CellID, e.Key.CellObjID)
		if _, ok := mapOcnResp[key]; !ok {
			mapOcnResp[key] = &pb.OcnRecord{
				OcnRecord: make(map[string]int32),
			}
		}
		innerKey := fmt.Sprintf("%s:%s:%s:%s", e.Value.Key.NodeID, e.Value.Key.PlmnID, e.Value.Key.CellID, e.Value.Key.CellObjID)
		value := e.Value.Value
		mapOcnResp[key].OcnRecord[innerKey] = int32(value)
	}

	return &pb.GetOcnResponse{
		OcnMap: mapOcnResp,
	}, nil
}
