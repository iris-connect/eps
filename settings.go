// IRIS Endpoint-Server (EPS)
// Copyright (C) 2021-2021 The IRIS Endpoint-Server Authors (see AUTHORS.md)
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package eps

type ChannelSettings struct {
	Name     string      `json:"name"`
	Type     string      `json:"type"`
	Services []string    `json:"services"`
	Settings interface{} `json:"settings"`
}

type DirectorySettings struct {
	Type     string      `json:"type"`
	Settings interface{} `json:"settings"`
}

type SigningSettings struct {
	Name                           string   `json:"name"`
	CACertificateFile              string   `json:"ca_certificate_file"`
	CAIntermediateCertificateFiles []string `json:"ca_intermediate_certificate_files"`
	CertificateFile                string   `json:"certificate_file"`
	KeyFile                        string   `json:"key_file"`
}

type MetricsSettings struct {
	BindAddress string `json:"bind_address"`
}

type Settings struct {
	Signing     *SigningSettings   `json:"signing"`
	Definitions *Definitions       `json:"definitions"`
	Channels    []*ChannelSettings `json:"channels"`
	Directory   *DirectorySettings `json:"directory"`
	Metrics     *MetricsSettings   `json:"metrics"`
	Name        string             `json:"name"`
}

type SettingsValidator func(settings map[string]interface{}) (interface{}, error)
