import {
  AmbientLight,
  BufferGeometry,
  CylinderGeometry,
  DirectionalLight,
  Float32BufferAttribute,
  LineBasicMaterial,
  LineSegments,
  Mesh,
  MeshStandardMaterial,
  PerspectiveCamera,
  Scene,
  SphereGeometry,
  Vector3,
  WebGLRenderer,
} from 'three'
import { OrbitControls } from 'three/examples/jsm/controls/OrbitControls.js'

import type { Point3, Segment, StructureGeometry } from './structureScene'

// 颜色与尺寸按“学生能看清”调过：球太大会糊成一团，太小又看不出堆积方式。
const BACKGROUND_COLOR = '#ffffff'
const CELL_COLOR = '#0b5d8a'
const BOND_COLOR = '#8a8f98'
const LIGHT_COLOR = '#ffffff'

const CAMERA_FIELD_OF_VIEW_DEGREE = 45
const CAMERA_DISTANCE_MARGIN = 1.15
const CAMERA_DIRECTION = { x: 1, y: 0.8, z: 1 }
const CAMERA_FAR_INITIAL_ANGSTROM = 1000
const MIN_CAMERA_NEAR_ANGSTROM = 0.01
const CAMERA_NEAR_DISTANCE_RATIO = 100
const CAMERA_FAR_DISTANCE_RATIO = 20
const DEGREE_TO_RADIAN = Math.PI / 180
const MAX_PIXEL_RATIO = 2

const AMBIENT_LIGHT_INTENSITY = 0.55
const KEY_LIGHT_INTENSITY = 1.1
const FILL_LIGHT_INTENSITY = 0.35
const KEY_LIGHT_OFFSET = { x: 1, y: 1.4, z: 1.6 }
const FILL_LIGHT_OFFSET = { x: -1.2, y: -0.6, z: -0.8 }

const SPHERE_WIDTH_SEGMENTS = 24
const SPHERE_HEIGHT_SEGMENTS = 16
const BOND_RADIAL_SEGMENTS = 10
const BOND_RADIUS_ANGSTROM = 0.12
const MATERIAL_ROUGHNESS = 0.6
const MATERIAL_METALNESS = 0.05
const BOND_ROUGHNESS = 0.7

// 常见元素的显示颜色（CPK 配色）；未收录的元素用固定调色板兜底。
const ELEMENT_COLORS: Record<string, string> = {
  H: '#e6e6e6',
  C: '#303030',
  N: '#3050f8',
  O: '#ff0d0d',
  F: '#90e050',
  Na: '#ab5cf2',
  Mg: '#8aff00',
  Al: '#bfa6a6',
  Si: '#f0c8a0',
  P: '#ff8000',
  S: '#ffff30',
  Cl: '#1ff01f',
  K: '#8f40d4',
  Ca: '#3dff00',
  Ti: '#bfc2c7',
  Cr: '#8a99c7',
  Fe: '#e06633',
  Co: '#f090a0',
  Ni: '#50d050',
  Cu: '#c88033',
  Zn: '#7d80b0',
  Ge: '#668f8f',
  Br: '#a62929',
  Mo: '#54b5b5',
  Ag: '#c0c0c0',
  I: '#940094',
  Cs: '#57178f',
  Ba: '#00c900',
  W: '#2194d6',
  Au: '#ffd123',
  Pb: '#575961',
}
const FALLBACK_ELEMENT_COLORS = ['#4e79a7', '#f28e2b', '#59a14f', '#b07aa1', '#76b7b2']

/**
 * 结果页的三维渲染器。
 *
 * 只负责把 `StructureGeometry` 画出来；化学结构怎么算不在这里。
 * 使用 Three.js 而不是依赖 `eval` 的渲染库，避免严格 CSP 下需要 `unsafe-eval`。
 */
export class StructureRenderer {
  private readonly host: HTMLElement
  private readonly geometry: StructureGeometry
  private readonly renderer: WebGLRenderer
  private readonly scene = new Scene()
  private readonly camera: PerspectiveCamera
  private readonly controls: OrbitControls
  private readonly resizeObserver: ResizeObserver
  private readonly disposables: Array<{ dispose(): void }> = []

  constructor(host: HTMLElement, geometry: StructureGeometry) {
    this.host = host
    this.geometry = geometry

    this.renderer = new WebGLRenderer({ antialias: true })
    this.renderer.setPixelRatio(Math.min(window.devicePixelRatio || 1, MAX_PIXEL_RATIO))
    this.renderer.setClearColor(BACKGROUND_COLOR, 1)
    this.renderer.domElement.setAttribute('aria-hidden', 'true')
    host.appendChild(this.renderer.domElement)

    const { width, height } = this.viewportSize()
    this.camera = new PerspectiveCamera(
      CAMERA_FIELD_OF_VIEW_DEGREE,
      width / height,
      MIN_CAMERA_NEAR_ANGSTROM,
      CAMERA_FAR_INITIAL_ANGSTROM,
    )
    this.controls = new OrbitControls(this.camera, this.renderer.domElement)
    this.controls.enableDamping = false
    this.controls.enablePan = false
    this.controls.addEventListener('change', () => this.render())

    this.addLights()
    this.addAtoms()
    this.addBonds()
    this.addCellEdges()
    this.resize()
    this.frameCamera()

    this.resizeObserver = new ResizeObserver(() => this.resize())
    this.resizeObserver.observe(host)
  }

  /** 回到刚打开结果页时的取景。 */
  resetView(): void {
    this.frameCamera()
  }

  /** 离开结果页时释放 WebGL 上下文、几何体与材质。 */
  dispose(): void {
    this.resizeObserver.disconnect()
    this.controls.dispose()
    for (const disposable of this.disposables) {
      disposable.dispose()
    }
    this.disposables.length = 0
    this.renderer.forceContextLoss()
    this.renderer.dispose()
    this.renderer.domElement.remove()
    this.scene.clear()
  }

  private addLights(): void {
    this.scene.add(new AmbientLight(LIGHT_COLOR, AMBIENT_LIGHT_INTENSITY))
    this.addDirectionalLight(KEY_LIGHT_OFFSET, KEY_LIGHT_INTENSITY)
    this.addDirectionalLight(FILL_LIGHT_OFFSET, FILL_LIGHT_INTENSITY)
  }

  private addDirectionalLight(offset: Point3, intensity: number): void {
    const { center, radius } = this.geometry
    const light = new DirectionalLight(LIGHT_COLOR, intensity)
    light.position.set(
      center.x + radius * offset.x,
      center.y + radius * offset.y,
      center.z + radius * offset.z,
    )
    light.target.position.set(center.x, center.y, center.z)
    this.scene.add(light, light.target)
  }

  private addAtoms(): void {
    const sphere = new SphereGeometry(1, SPHERE_WIDTH_SEGMENTS, SPHERE_HEIGHT_SEGMENTS)
    this.disposables.push(sphere)

    const materials = new Map<string, MeshStandardMaterial>()
    for (const atom of this.geometry.atoms) {
      let material = materials.get(atom.symbol)
      if (material === undefined) {
        material = new MeshStandardMaterial({
          color: colorForElement(atom.symbol),
          roughness: MATERIAL_ROUGHNESS,
          metalness: MATERIAL_METALNESS,
        })
        materials.set(atom.symbol, material)
        this.disposables.push(material)
      }

      const mesh = new Mesh(sphere, material)
      mesh.position.set(atom.position.x, atom.position.y, atom.position.z)
      mesh.scale.setScalar(this.geometry.atomRadius)
      this.scene.add(mesh)
    }
  }

  private addBonds(): void {
    if (this.geometry.bonds.length === 0) {
      return
    }

    const cylinder = new CylinderGeometry(1, 1, 1, BOND_RADIAL_SEGMENTS)
    const material = new MeshStandardMaterial({
      color: BOND_COLOR,
      roughness: BOND_ROUGHNESS,
      metalness: 0,
    })
    this.disposables.push(cylinder, material)

    const up = new Vector3(0, 1, 0)
    for (const bond of this.geometry.bonds) {
      this.addBond(cylinder, material, up, bond)
    }
  }

  private addBond(
    cylinder: CylinderGeometry,
    material: MeshStandardMaterial,
    up: Vector3,
    bond: Segment,
  ): void {
    const start = new Vector3(bond.start.x, bond.start.y, bond.start.z)
    const end = new Vector3(bond.end.x, bond.end.y, bond.end.z)
    const direction = end.clone().sub(start)
    const length = direction.length()
    if (length === 0) {
      return
    }

    const mesh = new Mesh(cylinder, material)
    mesh.position.copy(start).add(end).multiplyScalar(0.5)
    mesh.quaternion.setFromUnitVectors(up, direction.normalize())
    mesh.scale.set(BOND_RADIUS_ANGSTROM, length, BOND_RADIUS_ANGSTROM)
    this.scene.add(mesh)
  }

  private addCellEdges(): void {
    const positions: number[] = []
    for (const edge of this.geometry.cellEdges) {
      positions.push(
        edge.start.x,
        edge.start.y,
        edge.start.z,
        edge.end.x,
        edge.end.y,
        edge.end.z,
      )
    }

    const geometry = new BufferGeometry()
    geometry.setAttribute('position', new Float32BufferAttribute(positions, 3))
    const material = new LineBasicMaterial({ color: CELL_COLOR })
    this.disposables.push(geometry, material)
    this.scene.add(new LineSegments(geometry, material))
  }

  private frameCamera(): void {
    const { center, radius } = this.geometry
    const verticalFov = CAMERA_FIELD_OF_VIEW_DEGREE * DEGREE_TO_RADIAN
    const horizontalFov = 2 * Math.atan(Math.tan(verticalFov / 2) * this.camera.aspect)
    const fitFov = Math.min(verticalFov, horizontalFov)
    const distance = (radius / Math.sin(fitFov / 2)) * CAMERA_DISTANCE_MARGIN
    const direction = new Vector3(
      CAMERA_DIRECTION.x,
      CAMERA_DIRECTION.y,
      CAMERA_DIRECTION.z,
    ).normalize()

    this.camera.position.set(
      center.x + direction.x * distance,
      center.y + direction.y * distance,
      center.z + direction.z * distance,
    )
    this.camera.near = Math.max(
      MIN_CAMERA_NEAR_ANGSTROM,
      distance / CAMERA_NEAR_DISTANCE_RATIO,
    )
    this.camera.far = Math.max(this.camera.near * 2, distance * CAMERA_FAR_DISTANCE_RATIO)
    this.camera.updateProjectionMatrix()
    this.controls.target.set(center.x, center.y, center.z)
    this.camera.lookAt(center.x, center.y, center.z)
    this.controls.update()
    this.render()
  }

  private resize(): void {
    const { width, height } = this.viewportSize()
    this.renderer.setSize(width, height)
    this.camera.aspect = width / height
    this.camera.updateProjectionMatrix()
    this.render()
  }

  private viewportSize(): { width: number; height: number } {
    // 组件刚挂载时可能还没完成布局；ResizeObserver 会在拿到真实尺寸后再校正。
    return {
      width: Math.max(1, this.host.clientWidth),
      height: Math.max(1, this.host.clientHeight),
    }
  }

  private render(): void {
    this.renderer.render(this.scene, this.camera)
  }
}

function colorForElement(symbol: string): string {
  const known = ELEMENT_COLORS[symbol]
  if (known !== undefined) {
    return known
  }

  let hash = 0
  for (const character of symbol) {
    hash += character.charCodeAt(0)
  }
  return FALLBACK_ELEMENT_COLORS[hash % FALLBACK_ELEMENT_COLORS.length]
}
