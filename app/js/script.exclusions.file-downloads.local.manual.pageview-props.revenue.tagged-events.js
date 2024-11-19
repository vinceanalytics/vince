!function(){"use strict";var o=window.location,p=window.document,l=p.currentScript,s=l.getAttribute("data-api")||new URL(l.src).origin+"/api/event";function c(e,t){e&&console.warn("Ignoring Event: "+e),t&&t.callback&&t.callback()}function e(e,t){try{if("true"===window.localStorage.plausible_ignore)return c("localStorage flag",t)}catch(e){}var r=l&&l.getAttribute("data-include"),n=l&&l.getAttribute("data-exclude");if("pageview"===e){r=!r||r.split(",").some(a),n=n&&n.split(",").some(a);if(!r||n)return c("exclusion rule",t)}function a(e){return o.pathname.match(new RegExp("^"+e.trim().replace(/\*\*/g,".*").replace(/([^\.])\*/g,"$1[^\\s/]*")+"/?$"))}var r={},n=(r.n=e,r.u=t&&t.u?t.u:o.href,r.d=l.getAttribute("data-domain"),r.r=p.referrer||null,t&&t.meta&&(r.m=JSON.stringify(t.meta)),t&&t.props&&(r.p=t.props),t&&t.revenue&&(r.$=t.revenue),l.getAttributeNames().filter(function(e){return"event-"===e.substring(0,6)})),i=r.p||{},u=(n.forEach(function(e){var t=e.replace("event-",""),e=l.getAttribute(e);i[t]=i[t]||e}),r.p=i,new XMLHttpRequest);u.open("POST",s,!0),u.setRequestHeader("Content-Type","text/plain"),u.send(JSON.stringify(r)),u.onreadystatechange=function(){4===u.readyState&&t&&t.callback&&t.callback({status:u.status})}}var t=window.plausible&&window.plausible.q||[];window.plausible=e;for(var r=0;r<t.length;r++)e.apply(this,t[r]);function f(e){return e&&e.tagName&&"a"===e.tagName.toLowerCase()}var v=1;function n(e){var t,r,n,a;if("auxclick"!==e.type||e.button===v)return t=function(e){for(;e&&(void 0===e.tagName||!f(e)||!e.href);)e=e.parentNode;return e}(e.target),r=t&&t.href&&t.href.split("?")[0],!function e(t,r){if(!t||b<r)return!1;if(h(t))return!0;return e(t.parentNode,r+1)}(t,0)&&(n=r)&&(a=n.split(".").pop(),d.some(function(e){return e===a}))?m(e,t,{name:"File Download",props:{url:r}}):void 0}function m(e,t,r){var n,a=!1;function i(){a||(a=!0,window.location=t.href)}!function(e,t){if(!e.defaultPrevented)return t=!t.target||t.target.match(/^_(self|parent|top)$/i),e=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type,t&&e}(e,t)?((n={props:r.props}).revenue=r.revenue,plausible(r.name,n)):((n={props:r.props,callback:i}).revenue=r.revenue,plausible(r.name,n),setTimeout(i,5e3),e.preventDefault())}p.addEventListener("click",n),p.addEventListener("auxclick",n);var a=["pdf","xlsx","docx","txt","rtf","csv","exe","key","pps","ppt","pptx","7z","pkg","rar","gz","zip","avi","mov","mp4","mpeg","wmv","midi","mp3","wav","wma","dmg"],i=l.getAttribute("file-types"),u=l.getAttribute("add-file-types"),d=i&&i.split(",")||u&&u.split(",").concat(a)||a;function g(e){var e=h(e)?e:e&&e.parentNode,t={name:null,props:{},revenue:{}},r=e&&e.classList;if(r)for(var n=0;n<r.length;n++){var a,i,u=r.item(n),o=u.match(/plausible-event-(.+)(=|--)(.+)/),o=(o&&(a=o[1],i=o[3].replace(/\+/g," "),"name"==a.toLowerCase()?t.name=i:t.props[a]=i),u.match(/plausible-revenue-(.+)(=|--)(.+)/));o&&(a=o[1],i=o[3],t.revenue[a]=i)}return t}var b=3;function w(e){if("auxclick"!==e.type||e.button===v){for(var t,r,n,a,i=e.target,u=0;u<=b&&i;u++){if((n=i)&&n.tagName&&"form"===n.tagName.toLowerCase())return;f(i)&&(t=i),h(i)&&(r=i),i=i.parentNode}r&&(a=g(r),t?(a.props.url=t.href,m(e,t,a)):((e={}).props=a.props,e.revenue=a.revenue,plausible(a.name,e)))}}function h(e){var t=e&&e.classList;if(t)for(var r=0;r<t.length;r++)if(t.item(r).match(/plausible-event-name(=|--)(.+)/))return!0;return!1}p.addEventListener("submit",function(e){var t,r=e.target,n=g(r);function a(){t||(t=!0,r.submit())}n.name&&(e.preventDefault(),t=!1,setTimeout(a,5e3),(e={props:n.props,callback:a}).revenue=n.revenue,plausible(n.name,e))}),p.addEventListener("click",w),p.addEventListener("auxclick",w)}();