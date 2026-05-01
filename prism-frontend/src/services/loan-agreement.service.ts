import http from '@/services/http'
import { DaftarKegiatanService } from '@/services/daftar-kegiatan.service'
import type { ApiResponse, PaginatedResponse } from '@/types/api.types'
import type { DaftarKegiatan, DKProject } from '@/types/daftar-kegiatan.types'
import type {
  DKProjectLoanOption,
  LoanAgreement,
  LoanAgreementListParams,
  LoanAgreementPayload,
} from '@/types/loan-agreement.types'

function toDKProjectLoanOption(dk: DaftarKegiatan, project: DKProject): DKProjectLoanOption {
  const gbProjects = project.gb_projects ?? []
  const labelParts = [
    gbProjects.map((gb) => gb.gb_code).join(', '),
    project.project_name || project.objectives || project.program_title?.title || dk.subject,
  ].filter(Boolean)

  return {
    ...project,
    dk: project.dk ?? dk,
    dk_id: project.dk_id ?? dk.id,
    daftar_kegiatan_subject: dk.subject,
    gb_projects: gbProjects,
    locations: project.locations ?? [],
    financing_details: project.financing_details ?? [],
    loan_allocations: project.loan_allocations ?? [],
    activity_details: project.activity_details ?? [],
    label: labelParts.join(' - ') || dk.subject,
  }
}

export const LoanAgreementService = {
  async getLoanAgreements(params?: LoanAgreementListParams) {
    const response = await http.get<PaginatedResponse<LoanAgreement>>('/loan-agreements', {
      params,
    })
    return response.data
  },

  async getLoanAgreement(id: string) {
    const response = await http.get<ApiResponse<LoanAgreement>>(`/loan-agreements/${id}`)
    return response.data.data
  },

  async createLoanAgreement(data: LoanAgreementPayload) {
    const response = await http.post<ApiResponse<LoanAgreement>>('/loan-agreements', data)
    return response.data.data
  },

  async updateLoanAgreement(id: string, data: LoanAgreementPayload) {
    const response = await http.put<ApiResponse<LoanAgreement>>(`/loan-agreements/${id}`, data)
    return response.data.data
  },

  async deleteLoanAgreement(id: string) {
    await http.delete(`/loan-agreements/${id}`)
  },

  async getDKProjectOptions(search?: string) {
    const dkResponse = await DaftarKegiatanService.getDaftarKegiatan({ limit: 1000 })
    const optionGroups = await Promise.all(
      dkResponse.data.map(async (dk) => {
        const projectsResponse = await DaftarKegiatanService.getProjects(dk.id, { limit: 1000 })
        return projectsResponse.data.map((project) => toDKProjectLoanOption(dk, project))
      }),
    )
    const options = optionGroups.flat()
    const keyword = search?.trim().toLowerCase()

    if (!keyword) return options

    return options.filter((project) =>
      [
        project.label,
        project.project_name,
        project.objectives,
        project.program_title?.title,
        project.daftar_kegiatan_subject,
        project.gb_projects?.map((gb) => `${gb.gb_code} ${gb.project_name}`).join(' '),
      ]
        .filter(Boolean)
        .join(' ')
        .toLowerCase()
        .includes(keyword),
      )
  },

  async getDKProjectOption(dkId: string, projectId: string) {
    const [dk, project] = await Promise.all([
      DaftarKegiatanService.getDK(dkId),
      DaftarKegiatanService.getProject(dkId, projectId),
    ])

    return toDKProjectLoanOption(dk, project)
  },

  getAllowedLenderIds(project?: DKProject | null) {
    if (!project) return []
    return [
      ...new Set(
        project.financing_details
          .map((detail) => detail.lender?.id)
          .filter((id): id is string => Boolean(id)),
      ),
    ]
  },
}
