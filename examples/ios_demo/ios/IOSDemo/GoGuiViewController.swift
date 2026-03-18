import UIKit
import Metal
import QuartzCore

class GoGuiViewController: UIViewController {
    private var metalLayer: CAMetalLayer!
    private var displayLink: CADisplayLink?
    private var started = false

    override func loadView() {
        let v = UIView()
        v.backgroundColor = .black
        self.view = v
    }

    override func viewDidLoad() {
        super.viewDidLoad()

        guard let device = MTLCreateSystemDefaultDevice() else {
            fatalError("Metal not available")
        }

        metalLayer = CAMetalLayer()
        metalLayer.device = device
        metalLayer.pixelFormat = .bgra8Unorm
        metalLayer.contentsScale = UIScreen.main.scale
        metalLayer.framebufferOnly = true
        view.layer.addSublayer(metalLayer)

        let pan = UIPanGestureRecognizer(
            target: self, action: #selector(handlePan(_:)))
        view.addGestureRecognizer(pan)
    }

    override func viewDidLayoutSubviews() {
        super.viewDidLayoutSubviews()
        let bounds = view.bounds
        metalLayer.frame = bounds
        metalLayer.drawableSize = CGSize(
            width: bounds.width * UIScreen.main.scale,
            height: bounds.height * UIScreen.main.scale)

        let w = Int32(bounds.width)
        let h = Int32(bounds.height)
        let scale = Float(UIScreen.main.scale)

        if !started {
            let layerPtr = Unmanaged.passUnretained(metalLayer)
                .toOpaque()
            GoGuiStart(
                UnsafeMutableRawPointer(layerPtr),
                w, h, scale)
            started = true

            displayLink = CADisplayLink(
                target: self,
                selector: #selector(render))
            displayLink?.add(to: .main, forMode: .default)
        } else {
            GoGuiResize(w, h, scale)
        }
    }

    @objc private func render() {
        GoGuiRender()
    }

    @objc private func handlePan(
        _ gesture: UIPanGestureRecognizer
    ) {
        let loc = gesture.location(in: view)
        GoGuiTouchMoved(Float(loc.x), Float(loc.y))
    }

    override func touchesBegan(
        _ touches: Set<UITouch>, with event: UIEvent?
    ) {
        if let touch = touches.first {
            let loc = touch.location(in: view)
            GoGuiTouchBegan(Float(loc.x), Float(loc.y))
        }
    }

    override func touchesEnded(
        _ touches: Set<UITouch>, with event: UIEvent?
    ) {
        if let touch = touches.first {
            let loc = touch.location(in: view)
            GoGuiTouchEnded(Float(loc.x), Float(loc.y))
        }
    }

    override func touchesCancelled(
        _ touches: Set<UITouch>, with event: UIEvent?
    ) {
        if let touch = touches.first {
            let loc = touch.location(in: view)
            GoGuiTouchEnded(Float(loc.x), Float(loc.y))
        }
    }

    deinit {
        displayLink?.invalidate()
        GoGuiDestroy()
    }
}
